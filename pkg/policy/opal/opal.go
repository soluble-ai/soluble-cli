package opal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	policies "github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/policy/manager"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/opal"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type opalPolicies string

var Opal manager.PolicyType = opalPolicies("opal")

func (opalPolicies) GetName() string {
	return "opal"
}
func (opalPolicies) GetCode() string {
	return "opl"
}

func (opalPolicies) PreparePolicies(policies []*policies.Policy, dest string) error {
	for _, policy := range policies {
		for _, target := range policy.Targets {
			if err := preparePolicy(policy, target, dest); err != nil {
				return err
			}
		}
	}
	return nil
}

func preparePolicy(policy *policies.Policy, target policies.Target, dest string) error {
	rt, err := getPolicyText(policy, target)
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dest, fmt.Sprintf("%s.rego", policy.ID)))
	if err != nil {
		return err
	}
	defer f.Close()
	return rt.write(f, policy.Metadata)
}

func (opalPolicies) GetTestRunner(runOpts tools.RunOpts, target policies.Target) tools.Single {
	t := &opal.Tool{}
	t.RunOpts = runOpts
	return t
}

func (opalPolicies) ValidatePolicies(runOpts tools.RunOpts, policies []*policies.Policy) (validate manager.ValidateResult) {
	for _, policy := range policies {
		for _, target := range policy.Targets {
			policyRegoPath := filepath.Join(target.Path(policy), "policy.rego")
			if !util.FileExists(policyRegoPath) {
				validate.AppendError(
					fmt.Errorf("\"policy.rego\" is missing in %s", target.Path(policy)))
				continue
			}
			if policies.InputTypeForTarget[target] == "" {
				validate.AppendError(
					fmt.Errorf("opal does not support the %s target in %s", target, policy.Path))
				continue
			}
			_, err := getPolicyText(policy, target)
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

func getPolicyText(policy *policies.Policy, target policies.Target) (*policyText, error) {
	if td := policy.TargetData[target]; td != nil {
		return td.(*policyText), nil
	}
	rt, err := readPolicyText(filepath.Join(target.Path(policy), "policy.rego"))
	if err != nil {
		return nil, err
	}
	if rt.inputType == "" && target == policies.Terraform {
		// ok
	} else if rt.inputType != policies.InputTypeForTarget[target] {
		return nil, fmt.Errorf("%s must have input_type := \"%s\" for the %s target",
			policy.Path, policies.InputTypeForTarget[target], target)
	}
	policy.TargetData[target] = rt
	return rt, nil
}

func (opalPolicies) FindPolicyResult(findings assessments.Findings, id string) manager.PassFail {
	for _, f := range findings {
		if f.Tool != nil && f.Tool["policy_id"] == id {
			pass := f.Pass
			return &pass
		}
	}
	return nil
}

func init() {
	policies.RegisterPolicyType(Opal)
}
