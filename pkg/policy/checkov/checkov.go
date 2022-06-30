package checkov

import (
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
)

var supportedTargets = []policy.Target{
	policy.Terraform, policy.Cloudformation, policy.Kubernetes, policy.Docker,
	policy.Secrets,
}

func getTestRunner(m *policy.Manager, target policy.Target) tools.Single {
	t := &checkov.Tool{
		Framework: string(target),
	}
	t.RunOpts = m.RunOpts
	return t
}

func findRuleResult(findings assessments.Findings, id string) policy.PassFail {
	for _, finding := range findings {
		if finding.Tool["check_id"] == id {
			return &finding.Pass
		}
	}
	return nil
}
