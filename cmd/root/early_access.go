package root

import (
	"github.com/soluble-ai/soluble-cli/cmd/armscan"
	"github.com/soluble-ai/soluble-cli/cmd/cdkscan"
	"github.com/soluble-ai/soluble-cli/cmd/dockerfilescan"
	"github.com/soluble-ai/soluble-cli/cmd/policy"
	"github.com/soluble-ai/soluble-cli/cmd/tfplan"
	"github.com/soluble-ai/soluble-cli/cmd/tfplanscan"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/cloudmap"
	"github.com/spf13/cobra"
)

func earlyAccessCommand() *cobra.Command {
	c := &cobra.Command{
		Use:     "early-access",
		Aliases: []string{"ea"},
		Short:   "Early access commands",
		Long:    "Eary access commands may be incomplete, not fully working, and subject to change.",
		Hidden:  true,
	}
	c.AddCommand(
		policy.Command(),
		tfplan.Command(),
		armscan.Command(),
		cdkscan.Command(),
		tools.CreateCommand(&cloudmap.Tool{}),
		dockerfilescan.Command(),
		tfplanscan.Command(),
	)
	return c
}
