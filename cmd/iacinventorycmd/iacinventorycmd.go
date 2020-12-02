package iacinventorycmd

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/iacinventory"
	"github.com/spf13/cobra"
)

func createCommand(tool tools.Interface, use, short string) *cobra.Command {
	return tools.CreateCommand(tool, &cobra.Command{
		Use:   use,
		Short: short,
	})
}

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "iac-inventory",
		Short: "Inventory infrastructure-as-code and optionally send the inventory to Soluble",
	}
	c.AddCommand(
		createCommand(&iacinventory.Github{},
			"github", "Download and inventory github repositories"),
		createCommand(&iacinventory.Local{},
			"local", "Inventory local directories"))
	return c
}
