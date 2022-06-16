package helmscan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&checkov.Helm{})
	c.Use = "helm-scan"
	c.Short = "Scan helm charts"
	c.AddCommand(tools.CreateCommand(&checkov.Helm{}))
	return c
}
