package iacinventorycmd

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/iacinventory"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "iac-inventory",
		Short: "Look for infrastructure-as-code and optionally send the inventory to Soluble",
		Example: `  # run the inventory against github
  iac-inventory github --public

  # run the inventory against a local directory
  iac-invenory local --dir ./some/dir`,
	}
	c.AddCommand(githubCmd())
	c.AddCommand(localCmd())
	return c
}

func localCmd() *cobra.Command {
	return tools.CreateCommand(
		&iacinventory.FSIACInventoryScanner{},
		&cobra.Command{
			Use:   "local",
			Short: "IaC inventory for local directories",
			Args:  cobra.NoArgs,
		})
}

func githubCmd() *cobra.Command {
	return tools.CreateCommand(
		&iacinventory.GithubIacInventoryScanner{},
		&cobra.Command{
			Use:   "github",
			Short: "IaC inventory for GitHub",
			Args:  cobra.NoArgs,
		})
}
