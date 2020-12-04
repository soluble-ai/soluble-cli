package iacinventorycmd

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/iacinventory"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "iac-inventory",
		Short: "Inventory infrastructure-as-code and optionally send the inventory to Soluble",
	}
	c.AddCommand(
		tools.CreateCommand(&iacinventory.Github{}),
		tools.CreateCommand(&iacinventory.Local{}))
	return c
}
