package repoinventory

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/iacinventory"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return tools.CreateCommand(&iacinventory.Repo{})
}
