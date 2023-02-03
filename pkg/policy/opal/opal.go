package opal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/log"
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
	f, err := os.Create(filepath.Join(dest, fmt.Sprintf("%s-%s.rego", policy.ID, target)))
	if err != nil {
		return err
	}
	defer f.Close()
	return rt.write(f, policy.ID, target, policy.Metadata)
}

func (opalPolicies) GetTestRunner(runOpts tools.RunOpts, target policies.Target) tools.Single {
	return GetTestOpal(runOpts, target)
}

func GetTestOpal(runOpts tools.RunOpts, target policies.Target) tools.Single {
	toolPath, ok := os.LookupEnv("TEST_OPAL_TOOL_PATH")
	if ok {
		// TEST_OPAL_TOOL_PATH should be set to the binary location under the opal repo to run the tests with a local opal binary
		log.Infof("TEST_OPAL_TOOL_PATH=%s", toolPath)
		runOpts.ToolPath = toolPath
	}
	t := &opal.Tool{}
	switch target {
	case policies.ARM:
		t.InputType = "arm"
	case policies.Cloudformation:
		t.InputType = "cfn"
	case policies.Kubernetes:
		t.InputType = "k8s"
	case policies.Terraform:
		t.InputType = "tf"
	case policies.TerraformPlan:
		t.InputType = "tf-plan"
	}
	t.RunOpts = runOpts
	t.PrintResultOpt = true
	return t
}

func (opalPolicies) ValidatePolicies(runOpts tools.RunOpts, opalPolicies []*policies.Policy) (validate manager.ValidateResult) {
	for _, policy := range opalPolicies {
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

func (opalPolicies) FindPolicyResult(findings assessments.Findings, id string) []manager.PassFail {
	var passFail []manager.PassFail
	for _, f := range findings {
		log.Debugf("finding {info:%s} is {primary:%s}", f.Tool["policy_id"], f.Pass)
		if f.Tool != nil && f.Tool["policy_id"] == id {
			passFail = append(passFail, &f.Pass)
		}
	}
	return passFail
}

func init() {
	policies.RegisterPolicyType(Opal)
}
