package root

import (
	"github.com/soluble-ai/soluble-cli/cmd/policy"
	"github.com/soluble-ai/soluble-cli/cmd/tfplan"
	"github.com/spf13/cobra"
)

func earlyAccessCommand() *cobra.Command {
	c := &cobra.Command{
		Use:     "early-access",
		Aliases: []string{"ea"},
		Short:   "Alpha/pre-release commands subject to change",
		Hidden:  true,
	}
	c.AddCommand(
		policy.Command(),
		tfplan.Command(),
	)
	return c
}
