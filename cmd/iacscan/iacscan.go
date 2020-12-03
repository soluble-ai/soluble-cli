package iacscan

import (
	"fmt"

	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/all"
	cfnpythonlint "github.com/soluble-ai/soluble-cli/pkg/tools/cfn-python-lint"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/soluble-ai/soluble-cli/pkg/tools/secrets"
	"github.com/soluble-ai/soluble-cli/pkg/tools/terrascan"
	"github.com/soluble-ai/soluble-cli/pkg/tools/tfsec"
	"github.com/spf13/cobra"
)

func createCommand(tool tools.Interface) *cobra.Command {
	return tools.CreateCommand(tool,
		&cobra.Command{
			Use:   tool.Name(),
			Short: fmt.Sprintf("Scan infrastructure-as-code with %s", tool.Name()),
		})
}

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "iac-scan",
		Short: "Run an Infrastructure-as-code scanner",
		Example: `  # run the Infrastructure as Code scanner tool in the current directory if directory is not specified
  iac-scan <tool-name>
  iac-scan <tool-name> -d my-directory`,
	}
	terrascanTool := createCommand(&terrascan.Tool{})
	terrascanTool.Aliases = []string{"default"}
	secretsTool := createCommand(&secrets.Tool{})
	secretsTool.Short = "Scan infrastructure-as-code for secrets"
	c.AddCommand(
		terrascanTool,
		createCommand(&checkov.Tool{}),
		createCommand(&tfsec.Tool{}),
		createCommand(&cfnpythonlint.Tool{}),
		all.Command(),
		secretsTool,
	)

	// Disabling the cloudformation guard for now as it doesn't fit our strategy
	// c.AddCommand(createCommand(&cloudformationguard.Tool{}))
	return c
}
