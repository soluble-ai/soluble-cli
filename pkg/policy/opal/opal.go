package opal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/policy/manager"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/opal"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type opalRules string

var Opal manager.RuleType = opalRules("opal")

func (opalRules) GetName() string {
	return "opal"
}
func (opalRules) GetCode() string {
	return "opl"
}

func (opalRules) PrepareRules(rules []*policy.Rule, dest string) error {
	for _, rule := range rules {
		for _, target := range rule.Targets {
			if err := prepareRule(rule, target, dest); err != nil {
				return err
			}
		}
	}
	return nil
}

func prepareRule(rule *policy.Rule, target policy.Target, dest string) error {
	rt, err := getRuleText(rule, target)
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dest, fmt.Sprintf("%s.rego", rule.ID)))
	if err != nil {
		return err
	}
	defer f.Close()
	return rt.write(f, rule.Metadata)
}

func (opalRules) GetTestRunner(runOpts tools.RunOpts, target policy.Target) tools.Single {
	t := &opal.Tool{}
	t.RunOpts = runOpts
	return t
}

func (opalRules) ValidateRules(runOpts tools.RunOpts, rules []*policy.Rule) (validate manager.ValidateResult) {
	for _, rule := range rules {
		for _, target := range rule.Targets {
			ruleRegoPath := filepath.Join(target.Path(rule), "rule.rego")
			if !util.FileExists(ruleRegoPath) {
				validate.AppendError(
					fmt.Errorf("\"rule.rego\" is missing in %s", target.Path(rule)))
				continue
			}
			_, err := getRuleText(rule, target)
			if err != nil {
				validate.AppendError(err)
				validate.Invalid++
			} else {
				validate.Valid++
			}
		}
	}
	return
}

func getRuleText(rule *policy.Rule, target policy.Target) (*ruleText, error) {
	if td := rule.TargetData[target]; td != nil {
		return td.(*ruleText), nil
	}
	rt, err := readRuleText(filepath.Join(target.Path(rule), "rule.rego"))
	if err != nil {
		return nil, err
	}
	rule.TargetData[target] = rt
	return rt, nil
}

func (opalRules) FindRuleResult(findings assessments.Findings, id string) manager.PassFail {
	for _, f := range findings {
		if f.Tool != nil && f.Tool["rule_id"] == id {
			pass := f.Pass
			return &pass
		}
	}
	return nil
}

func init() {
	policy.RegisterRuleType(Opal)
}
