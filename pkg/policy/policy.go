package policy

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Rule struct {
	ID       string
	Path     string
	Metadata map[string]interface{}
	Targets  []Target
}

type Target string

const (
	Terraform      = Target("terraform")
	TerraformPlan  = Target("terraform-plan")
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
	PrepareRules(m *Manager, rules []*Rule, dest string) error
	ValidateRules(m *Manager, rules []*Rule) error
	GetTestRunner(m *Manager, target Target) tools.Single
	FindRuleResult(findings assessments.Findings, id string) PassFail
}

var allRuleTypes = map[string]RuleType{}

type Manager struct {
	tools.RunOpts
	Dir   string
	Rules map[RuleType][]*Rule
}

type TestMetrics struct {
	FailureCount int
	TestCount    int
}

type ValidateMetrics struct {
	Count int
}

func RegisterRuleType(ruleType RuleType) {
	allRuleTypes[ruleType.GetName()] = ruleType
}

func GetRuleType(ruleTypeName string) RuleType {
	return allRuleTypes[ruleTypeName]
}

func (m *Manager) Register(cmd *cobra.Command) {
	m.RunOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVarP(&m.Dir, "directory", "d", "", "Load policies from this directory")
	_ = cmd.MarkFlagRequired("directory")
}

func (m *Manager) LoadRules() error {
	m.Rules = make(map[RuleType][]*Rule)
	for _, ruleType := range allRuleTypes {
		if err := m.loadRules(ruleType); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) loadRules(ruleType RuleType) error {
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
		_, err := m.loadRule(ruleType, filepath.Join(ruleTypeDir, dirName))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) loadRule(ruleType RuleType, path string) (*Rule, error) {
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
	rule.Metadata["ruleId"] = rule.ID
	rule.Metadata["sid"] = rule.ID
	log.Debugf("Loaded %s from %s\n", rule.ID, rule.Path)
	m.Rules[ruleType] = append(m.Rules[ruleType], rule)
	return rule, nil
}

func (m *Manager) PrepareRules(dest string) error {
	var err error
	for ruleType, rules := range m.Rules {
		if perr := ruleType.PrepareRules(m, rules, dest); perr != nil {
			err = multierror.Append(err, perr)
		}
	}
	return err
}

func (m *Manager) ValidateRules() (ValidateMetrics, error) {
	var (
		metrics ValidateMetrics
		err     error
	)
	for _, ruleType := range m.getLoadedRuleTypes() {
		rules := m.Rules[ruleType]
		metrics.Count += len(rules)
		if verr := ruleType.ValidateRules(m, rules); verr != nil {
			err = multierror.Append(err, verr)
		}
	}
	return metrics, err
}

func (m *Manager) RuleCount() (count int) {
	for _, rules := range m.Rules {
		count += len(rules)
	}
	return
}

func (m *Manager) CreateTarBall(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	w := tar.NewWriter(gz)
	if err := m.writeRules(w); err != nil {
		return err
	}
	if err := m.writeUploadMetadata(w); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	log.Infof("Created tarball with {info:%d} rules", m.RuleCount())
	return gz.Close()
}

func (m *Manager) writeUploadMetadata(w *tar.Writer) error {
	env := xcp.GetCIEnv(m.Dir)
	dat, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return err
	}
	h := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "policies/upload.json",
		Size:     int64(len(dat)),
		ModTime:  time.Now(),
		Mode:     0644,
	}
	if err := w.WriteHeader(h); err != nil {
		return err
	}
	if _, err := w.Write(dat); err != nil {
		return err
	}
	return nil
}

func (m *Manager) writeRules(w *tar.Writer) error {
	for _, ruleType := range m.getLoadedRuleTypes() {
		rules := m.Rules[ruleType]
		for _, rule := range rules {
			log.Infof("Including {info:%s} from {primary:%s}", rule.ID, rule.Path)
			if err := m.writeRuleFiles(w, rule); err != nil {
				return err
			}
			if err := m.writeRuleMetadata(w, rule); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) writeRuleMetadata(w *tar.Writer, rule *Rule) error {
	dat, err := yaml.Marshal(rule.Metadata)
	if err != nil {
		return err
	}
	rpath, err := filepath.Rel(m.Dir, rule.Path)
	if err != nil {
		return err
	}
	h := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     fmt.Sprintf("%s/metadata.yaml", rpath),
		Size:     int64(len(dat)),
		ModTime:  time.Now(),
		Mode:     0644,
	}
	if err := w.WriteHeader(h); err != nil {
		return err
	}
	if _, err := w.Write(dat); err != nil {
		return err
	}
	return nil
}

func (m *Manager) writeRuleFiles(w *tar.Writer, rule *Rule) error {
	return filepath.Walk(rule.Path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "metadata.yaml" {
			return nil
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
}

func (m *Manager) getLoadedRuleTypes() (res []RuleType) {
	for ruleType := range m.Rules {
		res = append(res, ruleType)
	}
	// sort so that validate and test run in the same
	// order each time
	sort.Slice(res, func(i, j int) bool {
		return strings.Compare(res[i].GetName(), res[j].GetName()) > 0
	})
	return
}

func (m *Manager) TestRules() (TestMetrics, error) {
	metrics := TestMetrics{}
	dest, err := os.MkdirTemp("", "testrules*")
	if err != nil {
		return metrics, err
	}
	defer os.RemoveAll(dest)
	for ruleType, rules := range m.Rules {
		if err := ruleType.PrepareRules(m, rules, dest); err != nil {
			return metrics, err
		}
	}
	err = nil
	for _, ruleType := range m.getLoadedRuleTypes() {
		rules := m.Rules[ruleType]
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
		tool := ruleType.GetTestRunner(m, target)
		opts := tool.GetAssessmentOptions()
		opts.Tool = tool
		opts.DisableCustomPolicies = true
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
