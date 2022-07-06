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
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type checkovPython string

var CheckovPython policy.RuleType = checkovPython("checkov-py")

func (checkovPython) GetName() string {
	return "checkov-py"
}

func (checkovPython) GetCode() string {
	return "ckvpy"
}

func (checkovPython) PrepareRules(m *policy.Manager, rules []*policy.Rule, dst string) error {
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
	return name, nil
}

var checkAssignmentRe = regexp.MustCompile("^check *=")

func (h checkovPython) ValidateRules(m *policy.Manager, rules []*policy.Rule) error {
	var err error
	for _, rule := range rules {
		if e := h.validate(rule); e != nil {
			err = multierror.Append(err, e)
		}
	}
	if err != nil {
		return err
	}
	// We're going to run checkov against these rules just to
	// determine if they load w/o failure
	temp, err := os.MkdirTemp("", "checkov-py*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	policyDir := filepath.Join(temp, "policy")
	if err := os.MkdirAll(policyDir, 0777); err != nil {
		return err
	}
	if err := h.PrepareRules(m, rules, policyDir); err != nil {
		return err
	}
	t := &checkov.Tool{}
	t.Directory = temp
	t.UploadEnabled = false
	t.DisableCustomPolicies = true
	t.CustomPoliciesDir = policyDir
	t.RunOpts = m.RunOpts
	if err := t.Validate(); err != nil {
		return err
	}
	log.Infof("Verifying that checkov can load {info:checkov-py} rules")
	result, err := t.Run()
	if err != nil {
		return err
	}
	if result.ExecuteResult != nil {
		if strings.Contains(result.ExecuteResult.CombinedOutput, "Traceback") {
			// Look for individual rules
			for _, rule := range rules {
				if strings.Contains(result.ExecuteResult.CombinedOutput, rule.ID) {
					err = multierror.Append(fmt.Errorf("the python rule in %s does not load in checkov", rule.Path))
				}
			}
			if err == nil {
				err = fmt.Errorf("{info:checkov} has crashed with these custom {info:checkov-py} rules")
			}
		}
	}
	return err
}

func (h checkovPython) validate(rule *policy.Rule) error {
	var err error
	for _, target := range supportedTargets {
		rulePy := fmt.Sprintf("%s/%s/rule.py", rule.Path, target)
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
		if foundCheck {
			rule.Targets = append(rule.Targets, target)
		} else {
			err = multierror.Append(err, fmt.Errorf("%s did not contain an assignment to 'check'", rulePy))
		}
	}
	return err
}

func (checkovPython) GetTestRunner(m *policy.Manager, target policy.Target) tools.Single {
	return getTestRunner(m, target)
}

func (checkovPython) FindRuleResult(findings assessments.Findings, id string) policy.PassFail {
	return findRuleResult(findings, id)
}

func init() {
	policy.RegisterRuleType(CheckovPython)
}
