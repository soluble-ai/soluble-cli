package cfnscan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	cfnpythonlint "github.com/soluble-ai/soluble-cli/pkg/tools/cfn-python-lint"
	"github.com/soluble-ai/soluble-cli/pkg/tools/cfnnag"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&cfnpythonlint.Tool{})
	c.Use = "cloudformation-scan"
	c.Aliases = []string{"cfn-scan"}
	c.Short = "Scan cloudformation templates"
	c.Long = `Scan cloudformation templates with cfn-python-lint by default.

Use the sub-commands to explicitly choose a scanner to use.`
	c.AddCommand(
		tools.CreateCommand(&cfnnag.Tool{}),
		tools.CreateCommand(&cfnpythonlint.Tool{}),
		tools.CreateCommand(&checkov.Tool{
			Framework: "cloudformation",
		}),
	)
	return c
}
