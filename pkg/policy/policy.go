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
	None           = Target("")
)

type PassFail *bool

type RuleType interface {
	GetName() string
	GetCode() string
	PrepareRules(rules []*Rule, dest string) error
	Validate(rule *Rule) error
	GetTestRunner(target Target) tools.Single
	FindRuleResult(findings assessments.Findings, id string) PassFail
}

var allRuleTypes = map[string]RuleType{}

type Manager struct {
	Dir   string
	Rules map[RuleType][]*Rule
}

type TestMetrics struct {
	FailureCount int
	TestCount    int
}

func RegisterRuleType(ruleType RuleType) {
	allRuleTypes[ruleType.GetName()] = ruleType
}

func GetRuleType(ruleTypeName string) RuleType {
	return allRuleTypes[ruleTypeName]
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
	ruleTypeDir := filepath.Join(m.Dir, "policies", ruleType.GetName())
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
		ID:   id,
		Path: path,
	}
	if err := yaml.Unmarshal(d, &rule.Metadata); err != nil {
		return nil, fmt.Errorf("could not read %s - %w", rule.Path, err)
	}
	if rule.Metadata == nil {
		rule.Metadata = make(map[string]interface{})
	}
	rule.Metadata["id"] = rule.ID
	log.Debugf("Loaded %s from %s\n", rule.ID, rule.Path)
	m.Rules[ruleType] = append(m.Rules[ruleType], rule)
	return rule, nil
}

func (m *Manager) PrepareRules(dest string) error {
	var err error
	for ruleType, rules := range m.Rules {
		if perr := ruleType.PrepareRules(rules, dest); perr != nil {
			err = multierror.Append(err, perr)
		}
	}
	return err
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

func (m *Manager) TestRules() (TestMetrics, error) {
	metrics := TestMetrics{}
	dest, err := os.MkdirTemp("", "testrules*")
	if err != nil {
		return metrics, err
	}
	defer os.RemoveAll(dest)
	for ruleType, rules := range m.Rules {
		if err := ruleType.PrepareRules(rules, dest); err != nil {
			return metrics, err
		}
	}
	err = nil
	for ruleType, rules := range m.Rules {
		for _, rule := range rules {
			for _, target := range rule.Targets {
				if terr := m.testRuleTarget(&metrics, ruleType, rule, target, dest); terr != nil {
					err = multierror.Append(err, terr)
				}
			}
		}
	}
	return metrics, err
}

func (m *Manager) testRuleTarget(metrics *TestMetrics, ruleType RuleType, rule *Rule, target Target, dest string) error {
	failures := 0
tests:
	for _, passFailName := range []string{"pass", "fail"} {
		testDir := target.getTestsDir(rule, passFailName)
		if !util.DirExists(testDir) {
			continue
		}
		metrics.TestCount++
		tool := ruleType.GetTestRunner(target)
		opts := tool.GetAssessmentOptions()
		opts.Tool = tool
		opts.CustomPoliciesDir = dest
		opts.UploadEnabled = false
		if dir, ok := tool.(tools.HasDirectory); ok {
			dir.SetDirectory(testDir)
		}
		result, err := tools.RunSingleAssessment(tool)
		if err != nil {
			return err
		}
		passFailResult := ruleType.FindRuleResult(result.Findings, rule.ID)
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
				log.Infof("Policy {success:%s} %s %s - {success:OK}", p, passFailName, target)
			} else {
				log.Errorf("Policy {danger:%s} %s %s - {danger:FAILED}", p, passFailName, target)
				failures++
				metrics.FailureCount++
			}
			continue tests
		}
		log.Errorf("{primary:%s} - {danger:NOT FOUND}", testDir)
		failures++
		metrics.FailureCount++
	}
	if failures > 0 {
		return fmt.Errorf("%d tests have failed", failures)
	}
	return nil
}

func (t Target) getTestsDir(rule *Rule, passFailName string) string {
	if t != "" {
		return filepath.Join(rule.Path, string(t), "tests", passFailName)
	}
	return filepath.Join(rule.Path, "tests", passFailName)
}
