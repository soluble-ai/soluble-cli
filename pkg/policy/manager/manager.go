package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type PassFail *bool

type PolicyType interface {
	policy.PolicyType
	ValidatePolicies(runOpts tools.RunOpts, policies []*policy.Policy) ValidateResult
	GetTestRunner(runOpts tools.RunOpts, target policy.Target) tools.Single
	// Find a test result.  This must be tool-specific because the
	// findings have not been normalized.
	FindPolicyResult(findings assessments.Findings, id string) PassFail
}

type M struct {
	tools.RunOpts
	policy.Store
}

type TestMetrics struct {
	Policies []PolicyTestMetrics `json:"policies,omitempty"`
	Passed   int                 `json:"passed"`
	Failed   int                 `json:"failed"`
}

type PolicyTestMetrics struct {
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

type PolicyTemplate struct {
	PolicyType string
	CheckType  string
	PolicyName string
	PolicyDir  string
	// optional
	PolicyDesc     string
	PolicyTitle    string
	PolicySeverity string
	PolicyCategory string
	PolicyRsrcType string
}

var SeverityNames = util.NewStringSetWithValues([]string{
	"info", "low", "medium", "high", "critical",
})

func (pt *PolicyTemplate) ValidateCreateInput() error {
	// TODO add validation for optional input
	if isValid := regexp.MustCompile(`^[a-z0-9-]*$`).MatchString(pt.PolicyName); !isValid {
		return fmt.Errorf("invalid policy-name: %v. policy-name must consist only of [a-z0-9-]", pt.PolicyName)
	}

	pt.PolicyType = strings.ToLower(pt.PolicyType)
	if policy.GetRuleType(pt.PolicyType) == nil {
		return fmt.Errorf("invalid policy-type. policy-type is one of: %v", policy.ListRuleTypes())
	}

	pt.CheckType = strings.ToLower(pt.CheckType)
	if !policy.IsTarget(pt.CheckType) {
		return fmt.Errorf("invalid check-type. check-type is one of: %v", policy.ListTargets())
	}

	pt.PolicySeverity = strings.ToLower(pt.PolicySeverity)
	if !SeverityNames.Contains(pt.PolicySeverity) {
		return fmt.Errorf("invalid severity '%v'. severity is one of %v: ", pt.PolicySeverity, SeverityNames.Values())
	}

	if pt.PolicyDir == "policies" {
		if _, err := os.Stat(pt.PolicyDir); os.IsNotExist(err) {
			return fmt.Errorf("could not find '%v' directory in current directory."+
				"\ncreate 'policies' directory or use -d to target existing policies directory", pt.PolicyDir)
		}
	} else {
		dir := "/policies"
		if pt.PolicyDir[len(pt.PolicyDir)-len(dir):] != dir {
			return fmt.Errorf("invalid directory path: %v", pt.PolicyDir+
				"\nprovide path to existing policies directory")
		} else {
			if _, err := os.Stat(pt.PolicyDir); os.IsNotExist(err) {
				return fmt.Errorf("could not find directory: %v", pt.PolicyDir+
					"\ntarget existing policies directory.")
			}
		}
	}

	return nil
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

func (m *M) ValidatePolicies() ValidateResult {
	var result ValidateResult
	for _, policyType := range policy.GetPolicyTypes() {
		policies := m.Policies[policyType]
		log.Debugf("{primary:%s} has {info:%d} policies", policyType.GetName(), len(policies))
		if len(policies) == 0 {
			continue
		}
		typeResult := policyType.(PolicyType).ValidatePolicies(m.RunOpts, policies)
		if typeResult.Errors != nil {
			result.Errors = multierror.Append(result.Errors, typeResult.Errors)
		}
		result.Valid += typeResult.Valid
		result.Invalid += typeResult.Invalid
	}
	return result
}

func (m *M) TestPolicies() (TestMetrics, error) {
	metrics := TestMetrics{}
	dest, err := os.MkdirTemp("", "testpolicies*")
	if err != nil {
		return metrics, err
	}
	defer os.RemoveAll(dest)
	for policyType, policies := range m.Policies {
		if err := policyType.PreparePolicies(policies, dest); err != nil {
			return metrics, err
		}
	}
	err = nil
	for _, policyType := range policy.GetPolicyTypes() {
		policies := m.Policies[policyType]
		for _, policy := range policies {
			for _, target := range policy.Targets {
				if terr := m.testPolicyTarget(&metrics, policyType, policy, target, dest); terr != nil {
					err = multierror.Append(err, terr)
				}
			}
		}
	}
	return metrics, err
}

func (m *M) testPolicyTarget(metrics *TestMetrics, policyType policy.PolicyType, policy *policy.Policy, target policy.Target, dest string) error {
	mPolicyType, ok := policyType.(PolicyType)
	if !ok {
		return nil
	}
	failures := 0
tests:
	for _, passFailName := range []string{"pass", "fail"} {
		testDir := getTestsDir(target, policy, passFailName)
		if !util.DirExists(testDir) {
			continue
		}
		tool := mPolicyType.GetTestRunner(m.RunOpts, target)
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
		passFailResult := mPolicyType.FindPolicyResult(result.Findings, policy.ID)
		if passFailResult != nil {
			ok := *passFailResult
			if passFailName == "fail" {
				ok = !ok
			}
			p := policy.Path
			if rp, err := filepath.Rel(m.Dir, policy.Path); err == nil {
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
			metrics.Policies = append(metrics.Policies, PolicyTestMetrics{
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

func getTestsDir(t policy.Target, policy *policy.Policy, passFailName string) string {
	if t != "" {
		return filepath.Join(policy.Path, string(t), "tests", passFailName)
	}
	return filepath.Join(policy.Path, "tests", passFailName)
}
