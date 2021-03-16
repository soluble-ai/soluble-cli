package inventorycmd

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/iacinventory"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&iacinventory.Local{})
	c.Use = "inventory"
	c.AddCommand(tools.CreateCommand(&iacinventory.Local{}))
	return c
}
