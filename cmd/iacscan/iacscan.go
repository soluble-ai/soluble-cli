package iacscan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/all"
	cfnpythonlint "github.com/soluble-ai/soluble-cli/pkg/tools/cfn-python-lint"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/soluble-ai/soluble-cli/pkg/tools/secrets"
	"github.com/soluble-ai/soluble-cli/pkg/tools/semgrep"
	"github.com/soluble-ai/soluble-cli/pkg/tools/terrascan"
	"github.com/soluble-ai/soluble-cli/pkg/tools/tfsec"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "iac-scan",
		Short: "Run an Infrastructure-as-code scanner",
		Example: `  # run the Infrastructure as Code scanner tool in the current directory if directory is not specified
  iac-scan <tool-name>
  iac-scan <tool-name> -d my-directory`,
	}
	c.AddCommand(
		tools.CreateCommand(&terrascan.Tool{}),
		tools.CreateCommand(&checkov.Tool{}),
		tools.CreateCommand(&tfsec.Tool{}),
		tools.CreateCommand(&cfnpythonlint.Tool{}),
		tools.CreateCommand(&all.Tool{}),
		tools.CreateCommand(&secrets.Tool{}),
		tools.CreateCommand(&semgrep.Tool{}),
	)

	// Disabling the cloudformation guard for now as it doesn't fit our strategy
	// c.AddCommand(createCommand(&cloudformationguard.Tool{}))
	return c
}
