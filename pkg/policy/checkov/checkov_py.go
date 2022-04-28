package checkov

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/go-multierror"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type checkovPython string

var CheckovPython policy.RuleType = checkovPython("checkov-py")

func (checkovPython) GetName() string {
	return "checkov-py"
}

func (checkovPython) GetCode() string {
	return "ckv-py"
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
	return name, nil
}

var checkAssignmentRe = regexp.MustCompile("^check *=")

func (h checkovPython) Validate(rule *policy.Rule) error {
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

func (checkovPython) GetTestRunner(target policy.Target) tools.Single {
	return getTestRunner(target)
}

func (checkovPython) FindRuleResult(findings assessments.Findings, id string) policy.PassFail {
	return findRuleResult(findings, id)
}

func init() {
	policy.RegisterRuleType(CheckovPython)
}
