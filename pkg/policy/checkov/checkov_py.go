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

var CheckovPython manager.RuleType = checkovPython("checkov-py")

func (checkovPython) GetName() string {
	return "checkov-py"
}

func (checkovPython) GetCode() string {
	return "ckvpy"
}

func (checkovPython) PrepareRules(rules []*policy.Rule, dst string) error {
	var ruleFiles []string
	for _, rule := range rules {
		for _, target := range rule.Targets {
			name, err := preparePythonRule(rule, target, dst)
			if err != nil {
				return err
			}
			ruleFiles = append(ruleFiles, name)
		}
	}
	f, err := os.Create(filepath.Join(dst, "__init__.py"))
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "__all__ = [")
	for i, ruleFile := range ruleFiles {
		if i > 0 {
			fmt.Fprintf(f, ",")
		}
		fmt.Fprintf(f, " '%s'", ruleFile)
	}
	fmt.Fprintf(f, "]\n")
	return nil
}

func preparePythonRule(rule *policy.Rule, target policy.Target, dst string) (string, error) {
	s, err := os.Open(fmt.Sprintf("%s/%s/rule.py", rule.Path, target))
	if err != nil {
		return "", err
	}
	defer s.Close()
	name := fmt.Sprintf("%s-%s.py", rule.ID, target)
	d, err := os.Create(filepath.Join(dst, name))
	if err != nil {
		return "", err
	}
	defer d.Close()
	if _, err := io.Copy(d, s); err != nil {
		return "", err
	}
	fmt.Fprintf(d, "\ncheck.id = \"%s\"\n", rule.ID)
	fmt.Fprintf(d, "\ncheck.name = \"%s\"\n", quote(rule.Metadata.GetString("title")))
	return name, nil
}

var checkAssignmentRe = regexp.MustCompile("^check *=")

func (h checkovPython) ValidateRules(runOpts tools.RunOpts, rules []*policy.Rule) (validate manager.ValidateResult) {
	for _, rule := range rules {
		if e := h.validate(rule); e != nil {
			validate.Errors = multierror.Append(validate.Errors, e)
		}
	}
	if validate.Errors != nil {
		return
	}
	// We're going to run checkov against these rules just to
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
	if err := h.PrepareRules(rules, policyDir); err != nil {
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
	log.Infof("Verifying that checkov can load {info:checkov-py} rules")
	result, err := t.Run()
	if err != nil {
		validate.AppendError(err)
		return validate
	}
	validate.Valid = len(rules)
	if result.ExecuteResult != nil {
		if strings.Contains(result.ExecuteResult.CombinedOutput, "Traceback") {
			// Look for individual rules
			for _, rule := range rules {
				if strings.Contains(result.ExecuteResult.CombinedOutput, rule.ID) {
					err = multierror.Append(fmt.Errorf("the python rule in %s does not load in checkov", rule.Path))
					validate.Valid--
					validate.Invalid++
				}
			}
			if err == nil {
				err = fmt.Errorf("{info:checkov} has crashed with these custom {info:checkov-py} rules")
			}
			validate.AppendError(err)
		}
	}
	return
}

func (h checkovPython) validate(rule *policy.Rule) error {
	var err error
	for _, target := range rule.Targets {
		if verr := validateSupportedTarget(rule, target); err != nil {
			err = multierror.Append(err, verr)
		}
		rulePy := filepath.Join(target.Path(rule), "rule.py")
		if !util.FileExists(rulePy) {
			continue
		}
		foundCheck := false
		_ = util.ForEachLine(rulePy, func(line string) bool {
			if checkAssignmentRe.FindString(line) != "" {
				foundCheck = true
				return false
			}
			return true
		})
		if !foundCheck {
			err = multierror.Append(err, fmt.Errorf("%s did not contain an assignment to 'check'", rulePy))
		}
	}
	return err
}

func (checkovPython) GetTestRunner(runOpts tools.RunOpts, target policy.Target) tools.Single {
	return getTestRunner(runOpts, target)
}

func (checkovPython) FindRuleResult(findings assessments.Findings, id string) manager.PassFail {
	return findRuleResult(findings, id)
}

func init() {
	policy.RegisterRuleType(CheckovPython)
}
