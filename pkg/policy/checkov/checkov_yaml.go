package checkov

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/policy/manager"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"gopkg.in/yaml.v3"
)

type checkovYAML string

var CheckovYAML manager.PolicyType = checkovYAML("checkov")

func (checkovYAML) GetName() string {
	return "checkov"
}

func (checkovYAML) GetCode() string {
	return "ckv"
}

func (h checkovYAML) PreparePolicies(policies []*policy.Policy, dst string) error {
	for _, policy := range policies {
		for _, target := range policy.Targets {
			policyBody, err := h.readPolicy(policy, target)
			if err != nil {
				return err
			}
			util.GenericSet(&policyBody, "metadata/id", policy.ID)
			util.GenericSet(&policyBody, "metadata/name", policy.Metadata["title"])
			d, err := yaml.Marshal(policyBody)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(dst, fmt.Sprintf("%s-%s.yaml", target, policy.ID)), d, 0600); err != nil {
				return err
			}
		}
	}
	return nil
}

func (checkovYAML) readPolicy(policy *policy.Policy, target policy.Target) (map[string]interface{}, error) {
	d, err := os.ReadFile(filepath.Join(target.Path(policy), "policy.yaml"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var policyBody map[string]interface{}
	if err := yaml.Unmarshal(d, &policyBody); err != nil {
		return nil, fmt.Errorf("the YAML policy in %s/%s/policy.yaml is not legal yaml - %w", policy.Path, target, err)
	}
	return policyBody, nil
}

func (h checkovYAML) ValidatePolicies(runOpts tools.RunOpts, policies []*policy.Policy) (validate manager.ValidateResult) {
	for _, policy := range policies {
		if e := h.validate(policy); e != nil {
			validate.Invalid++
			validate.AppendError(e)
		} else {
			validate.Valid++
		}
	}
	return
}

func (h checkovYAML) validate(policy *policy.Policy) error {
	var err error
	for _, target := range policy.Targets {
		if verr := validateSupportedTarget(policy, target); verr != nil {
			err = multierror.Append(err, verr)
		}
		_, terr := h.readPolicy(policy, target)
		if terr != nil {
			err = multierror.Append(err, terr)
		}
	}
	return err
}

func (checkovYAML) GetTestRunner(runOpts tools.RunOpts, target policy.Target) tools.Single {
	return getTestRunner(runOpts, target)
}

func (checkovYAML) FindPolicyResult(findings assessments.Findings, id string) []manager.PassFail {
	return findPolicyResult(findings, id)
}

func init() {
	policy.RegisterPolicyType(CheckovYAML)
}
