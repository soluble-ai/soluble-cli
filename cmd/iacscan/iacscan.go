package iacscan

import (
	"fmt"

	"github.com/soluble-ai/soluble-cli/pkg/tools"
	cfnpythonlint "github.com/soluble-ai/soluble-cli/pkg/tools/cfn-python-lint"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/soluble-ai/soluble-cli/pkg/tools/semgrep"
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
	t := createCommand(&terrascan.Tool{})
	t.Aliases = []string{"default"}
	c.AddCommand(t)
	c.AddCommand(createCommand(&checkov.Tool{}))
	c.AddCommand(createCommand(&tfsec.Tool{}))
	c.AddCommand(createCommand(&cfnpythonlint.Tool{}))
	c.AddCommand(createCommand(&semgrep.Tool{}))

	// Disabling the cloudformation guard for now as it doesn't fit our strategy
	// c.AddCommand(createCommand(&cloudformationguard.Tool{}))
	return c
}
