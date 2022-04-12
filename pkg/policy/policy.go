package policy

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"gopkg.in/yaml.v3"
)

type Rule struct {
	Type     RuleType
	ID       string
	Path     string
	Metadata map[string]interface{}
	Targets  []Target
	Error    error
}

type Target string

const (
	Terraform      = Target("terraform")
	Cloudformation = Target("cloudformation")
	Kubernetes     = Target("kubernetes")
	Helm           = Target("helm")
	Docker         = Target("docker")
	Secrets        = Target("secrets")
)

var allTargets = []Target{
	Terraform, Cloudformation, Kubernetes, Helm, Docker, Secrets,
}

type PassFail *bool

type TestRunner interface {
	tools.Interface
}

type RuleType interface {
	GetCode() string
	Prepare(rule *Rule, target Target, dest string) error
	Validate(rule *Rule) error
	GetTestRunner(target Target) tools.Single
	FindRuleResult(findings assessments.Findings, id string) PassFail
}

var allRuleTypes = []RuleType{
	CheckovYAML,
}

type Manager struct {
	Dir   string
	Rules map[RuleType][]*Rule
}

func ruleTypeName(ruleType RuleType) string {
	return fmt.Sprint(ruleType)
}

func NewManager(dir string) *Manager {
	return &Manager{
		Dir:   dir,
		Rules: make(map[RuleType][]*Rule),
	}
}

func (m *Manager) LoadAllRules() error {
	for _, ruleType := range allRuleTypes {
		if err := m.LoadRules(ruleType); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) LoadRules(ruleType RuleType) error {
	ruleTypeDir := filepath.Join(m.Dir, "policies", ruleTypeName(ruleType))
	dirs, err := os.ReadDir(ruleTypeDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, ruleDir := range dirs {
		if !ruleDir.IsDir() {
			continue
		}
		dirName := ruleDir.Name()
		if dirName[0] == '.' {
			continue
		}
		_, err := m.LoadRule(ruleType, filepath.Join(ruleTypeDir, dirName))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return err
		}
	}
	return nil
}

func (m *Manager) LoadRule(ruleType RuleType, path string) (*Rule, error) {
	id := fmt.Sprintf("c-%s-%s", ruleType.GetCode(), strings.ReplaceAll(filepath.Base(path), "_", "-"))
	d, err := os.ReadFile(filepath.Join(path, "metadata.yaml"))
	if err != nil {
		return nil, err
	}
	rule := &Rule{
		Type: ruleType,
		ID:   id,
		Path: path,
	}
	for _, target := range allTargets {
		if util.DirExists(filepath.Join(rule.Path, string(target))) {
			rule.Targets = append(rule.Targets, target)
		}
	}
	if err := yaml.Unmarshal(d, &rule.Metadata); err != nil {
		return nil, fmt.Errorf("could not read %s - %w", rule.Path, err)
	}
	if rule.Metadata == nil {
		rule.Metadata = make(map[string]interface{})
	}
	rule.Metadata["id"] = rule.ID
	m.Rules[ruleType] = append(m.Rules[ruleType], rule)
	log.Debugf("Loaded %s from %s\n", rule.ID, rule.Path)
	return rule, nil
}

func (m *Manager) PrepareRules(dest string, ruleType RuleType, target Target) error {
	for _, rule := range m.Rules[ruleType] {
		if err := ruleType.Prepare(rule, target, dest); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) ValidateRules() error {
	var err error
	for ruleType, rules := range m.Rules {
		for _, rule := range rules {
			if rule.Error = ruleType.Validate(rule); rule.Error != nil {
				err = multierror.Append(err, rule.Error)
			}
		}
	}
	return err
}

func (m *Manager) CreateTarBall(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	w := tar.NewWriter(f)
	err = filepath.Walk(m.Dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rpath, err := filepath.Rel(m.Dir, path)
		if err != nil {
			return err
		}
		h := &tar.Header{
			Typeflag: tar.TypeReg,
			Name:     rpath,
			Size:     info.Size(),
			ModTime:  info.ModTime(),
			Mode:     0644,
		}
		if base := filepath.Base(path); base[0] == '.' {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			h.Typeflag = tar.TypeDir
			h.Mode = 0755
		}
		if err := w.WriteHeader(h); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(w, f); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return w.Close()
}

func (m *Manager) TestRules() error {
	var err error
	for ruleType := range m.Rules {
		terr := m.TestRuleType(ruleType)
		if terr != nil {
			err = multierror.Append(err, terr)
		}
	}
	return err
}

func (m *Manager) TestRuleType(ruleType RuleType) error {
	var err error
	for _, rule := range m.Rules[ruleType] {
		terr := m.TestRule(rule)
		if terr != nil {
			err = multierror.Append(err, terr)
		}
	}
	return err
}

func (m *Manager) TestRule(rule *Rule) error {
	var err error
	for _, target := range rule.Targets {
		terr := m.TestRuleTarget(rule, target)
		if terr != nil {
			err = multierror.Append(err, terr)
		}
	}
	return err
}

func (m *Manager) TestRuleTarget(rule *Rule, target Target) error {
	dir, err := os.MkdirTemp("", "test*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	if err := rule.Type.Prepare(rule, target, dir); err != nil {
		return err
	}
	failures := false
tests:
	for _, passFailName := range []string{"pass", "fail"} {
		testDir := filepath.Join(rule.Path, string(target), "tests", passFailName)
		if !util.DirExists(testDir) {
			continue
		}
		tool := rule.Type.GetTestRunner(target)
		opts := tool.GetAssessmentOptions()
		opts.Tool = tool
		opts.CustomPoliciesDir = dir
		opts.UploadEnabled = false
		if dir, ok := tool.(tools.HasDirectory); ok {
			dir.SetDirectory(testDir)
		}
		opts.Quiet = true
		result, err := tools.RunSingleAssessment(tool)
		if err != nil {
			return err
		}
		passFailResult := rule.Type.FindRuleResult(result.Findings, rule.ID)
		if passFailResult != nil {
			ok := *passFailResult
			if passFailName == "fail" {
				ok = !ok
			}
			p := rule.Path
			if rp, err := filepath.Rel(m.Dir, rule.Path); err == nil {
				p = rp
			}
			if ok {
				log.Infof("{primary:%s} %s %s - {info:OK}", p, passFailName, target)
			} else {
				log.Errorf("{primary:%s} %s %s - {danger:FAILED}", p, passFailName, target)
				failures = true
			}
			continue tests
		}
		log.Errorf("{primary:%s} - {danger:NOT FOUND}", testDir)
		failures = true
	}
	if failures {
		return fmt.Errorf("one or more tests have failed")
	}
	return nil
}
