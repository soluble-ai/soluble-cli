package all

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return tools.CreateCommand(&Tool{},
		&cobra.Command{
			Use:   "all",
			Short: "Find infrastructure-as-code and scan with recommended tools",
			Long: `Find infrastructure-as-code and scan with the following tools:

Cloudformation templates - cfn-python-lint
Terraform                - checkov
Kuberentes manifests     - checkov
Everything               - secrets			
`,
		},
	)
}
