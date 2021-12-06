package tfplan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/tfscore"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:     "terraform-plan",
		Short:   "Scan or generate terraform plans",
		Aliases: []string{"tf-plan"},
	}
	c.AddCommand(
		tools.CreateCommand(&tfscore.Tool{}),
		tools.CreateCommand(&tfscore.PlanTool{}),
	)
	return c
}
