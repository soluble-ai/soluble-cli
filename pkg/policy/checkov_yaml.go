package policy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"gopkg.in/yaml.v3"
)

type checkovYAMLType string

var CheckovYAML RuleType = checkovYAMLType("checkov")

func (checkovYAMLType) GetCode() string {
	return "ckv"
}

func (h checkovYAMLType) Prepare(rule *Rule, target Target, dst string) error {
	ruleBody, err := h.readRule(rule, target)
	if err != nil {
		return err
	}
	util.GenericSet(&ruleBody, "metadata/id", rule.ID)
	d, err := yaml.Marshal(ruleBody)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dst, fmt.Sprintf("%s-%s.yaml", target, rule.ID)), d, 0600)
}

func (checkovYAMLType) readRule(rule *Rule, target Target) (map[string]interface{}, error) {
	d, err := os.ReadFile(filepath.Join(rule.Path, string(target), "rule.yaml"))
	if err != nil {
		return nil, err
	}
	var ruleBody map[string]interface{}
	if err := yaml.Unmarshal(d, &ruleBody); err != nil {
		return nil, fmt.Errorf("the YAML rule in %s/%s/rule.yaml is not legal yaml - %w", rule.Path, target, err)
	}
	return ruleBody, nil
}

func (h checkovYAMLType) Validate(rule *Rule) error {
	var err error
	for _, target := range rule.Targets {
		_, terr := h.readRule(rule, target)
		if terr != nil {
			err = multierror.Append(err, terr)
		}
	}
	return err
}

func (checkovYAMLType) GetTestRunner(target Target) tools.Single {
	return &checkov.Tool{
		Framework: string(target),
	}
}

func (checkovYAMLType) FindRuleResult(findings assessments.Findings, id string) PassFail {
	for _, finding := range findings {
		if finding.Tool["check_id"] == id {
			return &finding.Pass
		}
	}
	return nil
}
