package depscan

import (
	"os"

	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/trivy"
	"github.com/spf13/cobra"
)

type DepScanOpts struct {
	options.PrintClientOpts
}

func (opts *DepScanOpts) Register(c *cobra.Command) {
	opts.PrintClientOpts.Register(c)
}

func Command() *cobra.Command {
	return depScanCommand()
}

func depScanCommand() *cobra.Command {
	opts := &DepScanOpts{}
	c := &cobra.Command{
		Use:   "dep-scan",
		Short: "Scan dependencies in the local directory",
		Args:  cobra.NoArgs,
	}
	pwd, _ := os.Getwd()
	c.AddCommand(tools.CreateCommand(&trivy.DirTool{
		Dir: pwd,
	}))
	opts.Register(c)
	return c
}
