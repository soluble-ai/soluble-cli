package manager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type PassFail *bool

type RuleType interface {
	policy.RuleType
	ValidateRules(runOpts tools.RunOpts, rules []*policy.Rule) ValidateResult
	GetTestRunner(runOpts tools.RunOpts, target policy.Target) tools.Single
	// Find a test result.  This must be tool-specific because the
	// findings have not been normalized.
	FindRuleResult(findings assessments.Findings, id string) PassFail
}

type M struct {
	tools.RunOpts
	policy.Store
}

type TestMetrics struct {
	Rules  []RuleTestMetrics `json:"rules,omitempty"`
	Passed int               `json:"passed"`
	Failed int               `json:"failed"`
}

type RuleTestMetrics struct {
	Path     string        `json:"path"`
	Target   policy.Target `json:"target"`
	TestType string        `json:"test_type"`
	Success  bool          `json:"success"`
}

type ValidateResult struct {
	Errors  error `json:"-"`
	Valid   int   `json:"valid"`
	Invalid int   `json:"invalid"`
}

func (vr *ValidateResult) AppendError(err error) {
	vr.Errors = multierror.Append(vr.Errors, err)
}

func (m *M) Register(cmd *cobra.Command) {
	m.RunOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVarP(&m.Dir, "directory", "d", "", "Load policies from this directory")
	_ = cmd.MarkFlagRequired("directory")
}

func (m *M) ValidateRules() ValidateResult {
	var result ValidateResult
	for _, ruleType := range policy.GetRuleTypes() {
		rules := m.Rules[ruleType]
		log.Debugf("{primary:%s} has {info:%d} rules", ruleType.GetName(), len(rules))
		if len(rules) == 0 {
			continue
		}
		typeResult := ruleType.(RuleType).ValidateRules(m.RunOpts, rules)
		if typeResult.Errors != nil {
			result.Errors = multierror.Append(result.Errors, typeResult.Errors)
		}
		result.Valid += typeResult.Valid
		result.Invalid += typeResult.Invalid
	}
	return result
}

func (m *M) TestRules() (TestMetrics, error) {
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
	for _, ruleType := range policy.GetRuleTypes() {
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

func (m *M) testRuleTarget(metrics *TestMetrics, ruleType policy.RuleType, rule *policy.Rule, target policy.Target, dest string) error {
	mRuleType, ok := ruleType.(RuleType)
	if !ok {
		return nil
	}
	failures := 0
tests:
	for _, passFailName := range []string{"pass", "fail"} {
		testDir := getTestsDir(target, rule, passFailName)
		if !util.DirExists(testDir) {
			continue
		}
		tool := mRuleType.GetTestRunner(m.RunOpts, target)
		opts := tool.GetAssessmentOptions()
		opts.Tool = tool
		opts.DisableCustomPolicies = true
		opts.PreparedCustomPoliciesDir = dest
		opts.UploadEnabled = false
		if dir, ok := tool.(tools.HasDirectory); ok {
			dir.SetDirectory(testDir)
		}
		result, err := tools.RunSingleAssessment(tool)
		if err != nil {
			return err
		}
		passFailResult := mRuleType.FindRuleResult(result.Findings, rule.ID)
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
				metrics.Passed++
			} else {
				log.Errorf("Policy {danger:%s} %s %s - {danger:FAILED}", p, passFailName, target)
				failures++
				metrics.Failed++
			}
			metrics.Rules = append(metrics.Rules, RuleTestMetrics{
				Path:     p,
				Target:   target,
				TestType: passFailName,
				Success:  ok,
			})
			continue tests
		}
		log.Errorf("{primary:%s} - {danger:NOT FOUND}", testDir)
		metrics.Failed++
		failures++
	}
	if failures > 0 {
		return fmt.Errorf("%d tests have failed", failures)
	}
	return nil
}

func getTestsDir(t policy.Target, rule *policy.Rule, passFailName string) string {
	if t != "" {
		return filepath.Join(rule.Path, string(t), "tests", passFailName)
	}
	return filepath.Join(rule.Path, "tests", passFailName)
}
