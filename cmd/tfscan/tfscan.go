package tfscan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/soluble-ai/soluble-cli/pkg/tools/terrascan"
	"github.com/soluble-ai/soluble-cli/pkg/tools/tfsec"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&checkov.Tool{
		Framework: "terraform",
	})
	c.Use = "terraform-scan"
	c.Aliases = []string{"tf-scan"}
	c.Short = `Scan terraform with checkov`
	c.Long = `Scan terraform with checkov by default.

Use a sub-command to explicitly choose a scanner.`
	c.AddCommand(
		tools.CreateCommand(&tfsec.Tool{}),
		tools.CreateCommand(&terrascan.Tool{}),
		tools.CreateCommand(&checkov.Tool{
			Framework: "terraform",
		}),
	)
	return c
}
