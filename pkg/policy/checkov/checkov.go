package checkov

import (
	"fmt"
	"regexp"

	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/policy/manager"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
)

var supportedTargets = []policy.Target{
	policy.Terraform, policy.Cloudformation, policy.Kubernetes, policy.Docker,
	policy.Secrets,
}

func validateSupportedTarget(policy *policy.Policy, target policy.Target) error {
	for _, st := range supportedTargets {
		if target == st {
			return nil
		}
	}
	return fmt.Errorf("the policy %s contains an unsupported target %s", policy.Path, target)
}

func getTestRunner(runOpts tools.RunOpts, target policy.Target) tools.Single {
	t := &checkov.Tool{
		Framework: string(target),
	}
	t.RunOpts = runOpts
	return t
}

func findPolicyResult(findings assessments.Findings, id string) manager.PassFail {
	for _, finding := range findings {
		if finding.Tool["check_id"] == id {
			return &finding.Pass
		}
	}
	return nil
}

var quoteRegexp = regexp.MustCompile(`(["\\])`)

func quote(s string) string {
	return quoteRegexp.ReplaceAllString(s, `\$1`)
}
