package iacinventorycmd

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/iacinventory"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return tools.CreateCommand(&iacinventory.GithubIacInventoryScanner{},
		&cobra.Command{
			Use:   "iac-inventory",
			Short: "Look for infrastructure-as-code and optionally send the inventory to Soluble",
		})
}
