package depscan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/retirejs"
	"github.com/soluble-ai/soluble-cli/pkg/tools/trivyfs"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&trivyfs.Tool{})
	c.Use = "dep-scan"
	c.Short = "Scan application dependencies"
	c.Long = `Scan application dependencies with trivy by default`
	c.AddCommand(tools.CreateCommand(&trivyfs.Tool{}))
	c.AddCommand(tools.CreateCommand(&retirejs.Tool{}))
	return c
}
