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

type PolicyType interface {
	policy.PolicyType
	ValidatePolicies(runOpts tools.RunOpts, policies []*policy.Policy) ValidateResult
	GetTestRunner(runOpts tools.RunOpts, target policy.Target) tools.Single
	// Find a test result.  This must be tool-specific because the
	// findings have not been normalized.
	FindPolicyResult(findings assessments.Findings, id string) []PassFail
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
	TestPath string        `json:"test_path"`
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
}

func (m *M) RegisterDownload(cmd *cobra.Command) {
	m.RunOpts.Register(cmd)
}

func (m *M) RegisterUpload(cmd *cobra.Command) {
	m.RunOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVarP(&m.Dir, "directory", "d", "", "Load policies from this directory. Path should point to the parent directory of your /policies directory.")
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
		if result.ExecuteResult != nil && result.ExecuteResult.FailureType != tools.NoFailure {
			log.Errorf("Failed to execute policy engine - {danger:%s}\n%s", result.ExecuteResult.FailureMessage, result.ExecuteResult.CombinedOutput)
			return fmt.Errorf("policy engine could not run")
		}
		passFailResults := mPolicyType.FindPolicyResult(result.Findings, policy.ID)
		if len(passFailResults) == 0 {
			log.Errorf("{primary:%s} - {danger:no findings with id %s found in results}", testDir, policy.ID)
			log.Errorf("%s", result.Findings)
			metrics.Failed++
			failures++
		}

		for i, passFailResult := range passFailResults {
			if passFailResult != nil {
				ok := *passFailResult
				if passFailName == "fail" {
					ok = !ok
				}
				p := policy.Path
				testPath := result.Findings[i].RepoPath
				testFile := filepath.Base(testPath)
				if rp, err := filepath.Rel(m.Dir, policy.Path); err == nil {
					p = rp
				}
				if ok {
					log.Infof("Policy {success:%s} test {success:%s} %s %s - {success:OK}", p, testFile, passFailName, target)
					metrics.Passed++
				} else {
					log.Errorf("Policy {danger:%s} test {danger:%s} %s %s - {danger:FAILED}", p, testFile, passFailName, target)
					failures++
					metrics.Failed++
				}
				metrics.Policies = append(metrics.Policies, PolicyTestMetrics{
					Path:     p,
					Target:   target,
					TestPath: testPath,
					TestType: passFailName,
					Success:  ok,
				})
			}
		}
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
