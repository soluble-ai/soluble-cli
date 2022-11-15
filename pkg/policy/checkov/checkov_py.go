package checkov

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/policy/manager"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type checkovPython string

var CheckovPython manager.PolicyType = checkovPython("checkov-py")

func (checkovPython) GetName() string {
	return "checkov-py"
}

func (checkovPython) GetCode() string {
	return "ckvpy"
}

func (checkovPython) PreparePolicies(policies []*policy.Policy, dst string) error {
	var policyFiles []string
	for _, policy := range policies {
		for _, target := range policy.Targets {
			name, err := preparePythonPolicy(policy, target, dst)
			if err != nil {
				return err
			}
			policyFiles = append(policyFiles, name)
		}
	}
	f, err := os.Create(filepath.Join(dst, "__init__.py"))
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "__all__ = [")
	for i, policyFile := range policyFiles {
		if i > 0 {
			fmt.Fprintf(f, ",")
		}
		fmt.Fprintf(f, " '%s'", policyFile)
	}
	fmt.Fprintf(f, "]\n")
	return nil
}

func preparePythonPolicy(policy *policy.Policy, target policy.Target, dst string) (string, error) {
	s, err := os.Open(fmt.Sprintf("%s/%s/policy.py", policy.Path, target))
	if err != nil {
		return "", err
	}
	defer s.Close()
	name := fmt.Sprintf("%s-%s.py", policy.ID, target)
	d, err := os.Create(filepath.Join(dst, name))
	if err != nil {
		return "", err
	}
	defer d.Close()
	if _, err := io.Copy(d, s); err != nil {
		return "", err
	}
	fmt.Fprintf(d, "\ncheck.id = \"%s\"\n", policy.ID)
	fmt.Fprintf(d, "\ncheck.name = \"%s\"\n", quote(policy.Metadata.GetString("title")))
	return name, nil
}

var checkAssignmentRe = regexp.MustCompile("^check *=")

func (h checkovPython) ValidatePolicies(runOpts tools.RunOpts, policies []*policy.Policy) (validate manager.ValidateResult) {
	for _, policy := range policies {
		if e := h.validate(policy); e != nil {
			validate.Errors = multierror.Append(validate.Errors, e)
		}
	}
	if validate.Errors != nil {
		return
	}
	// We're going to run checkov against these policies just to
	// determine if they load w/o failure
	temp, err := os.MkdirTemp("", "checkov-py*")
	if err != nil {
		validate.AppendError(err)
		return
	}
	defer os.RemoveAll(temp)
	policyDir := filepath.Join(temp, "policy")
	if err := os.MkdirAll(policyDir, 0777); err != nil {
		validate.AppendError(err)
		return
	}
	if err := h.PreparePolicies(policies, policyDir); err != nil {
		validate.AppendError(err)
		return
	}
	t := &checkov.Tool{}
	t.Directory = temp
	t.UploadEnabled = false
	t.DisableCustomPolicies = true
	t.CustomPoliciesDir = policyDir
	t.RunOpts = runOpts
	if err := t.Validate(); err != nil {
		validate.AppendError(err)
		return
	}
	log.Infof("Verifying that checkov can load {info:checkov-py} policies")
	result, err := t.Run()
	if err != nil {
		validate.AppendError(err)
		return validate
	}
	validate.Valid = len(policies)
	if result.ExecuteResult != nil {
		if strings.Contains(result.ExecuteResult.CombinedOutput, "Traceback") {
			// Look for individual policies
			for _, policy := range policies {
				if strings.Contains(result.ExecuteResult.CombinedOutput, policy.ID) {
					err = multierror.Append(fmt.Errorf("the python policy in %s does not load in checkov", policy.Path))
					validate.Valid--
					validate.Invalid++
				}
			}
			if err == nil {
				err = fmt.Errorf("{info:checkov} has crashed with these custom {info:checkov-py} policies")
			}
			validate.AppendError(err)
		}
	}
	return
}

func (h checkovPython) validate(policy *policy.Policy) error {
	var err error
	for _, target := range policy.Targets {
		if verr := validateSupportedTarget(policy, target); err != nil {
			err = multierror.Append(err, verr)
		}
		policyPy := filepath.Join(target.Path(policy), "policy.py")
		if !util.FileExists(policyPy) {
			continue
		}
		foundCheck := false
		_ = util.ForEachLine(policyPy, func(line string) bool {
			if checkAssignmentRe.FindString(line) != "" {
				foundCheck = true
				return false
			}
			return true
		})
		if !foundCheck {
			err = multierror.Append(err, fmt.Errorf("%s did not contain an assignment to 'check'", policyPy))
		}
	}
	return err
}

func (checkovPython) GetTestRunner(runOpts tools.RunOpts, target policy.Target) tools.Single {
	return getTestRunner(runOpts, target)
}

func (checkovPython) FindPolicyResult(findings assessments.Findings, id string) []manager.PassFail {
	return findPolicyResult(findings, id)
}

func init() {
	policy.RegisterPolicyType(CheckovPython)
}
