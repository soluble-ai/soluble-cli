package checkov

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"gopkg.in/yaml.v3"
)

type checkovYAML string

var CheckovYAML policy.RuleType = checkovYAML("checkov")

func (checkovYAML) GetName() string {
	return "checkov"
}

func (checkovYAML) GetCode() string {
	return "ckv"
}

func (h checkovYAML) PrepareRules(rules []*policy.Rule, dst string) error {
	for _, rule := range rules {
		for _, target := range rule.Targets {
			ruleBody, err := h.readRule(rule, target)
			if err != nil {
				return err
			}
			util.GenericSet(&ruleBody, "metadata/id", rule.ID)
			d, err := yaml.Marshal(ruleBody)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(dst, fmt.Sprintf("%s-%s.yaml", target, rule.ID)), d, 0600); err != nil {
				return err
			}
		}
	}
	return nil
}

func (checkovYAML) readRule(rule *policy.Rule, target policy.Target) (map[string]interface{}, error) {
	d, err := os.ReadFile(filepath.Join(rule.Path, string(target), "rule.yaml"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var ruleBody map[string]interface{}
	if err := yaml.Unmarshal(d, &ruleBody); err != nil {
		return nil, fmt.Errorf("the YAML rule in %s/%s/rule.yaml is not legal yaml - %w", rule.Path, target, err)
	}
	return ruleBody, nil
}

func (h checkovYAML) Validate(rule *policy.Rule) error {
	var err error
	for _, target := range supportedTargets {
		body, terr := h.readRule(rule, target)
		if terr != nil {
			err = multierror.Append(err, terr)
		}
		if body != nil {
			rule.Targets = append(rule.Targets, target)
		}
	}
	return err
}

func (checkovYAML) GetTestRunner(target policy.Target) tools.Single {
	return getTestRunner(target)
}

func (checkovYAML) FindRuleResult(findings assessments.Findings, id string) policy.PassFail {
	return findRuleResult(findings, id)
}

func init() {
	policy.RegisterRuleType(CheckovYAML)
}
