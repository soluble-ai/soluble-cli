package depscan

import (
	"fmt"
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
	var scanTools []string
	opts := &DepScanOpts{}
	c := &cobra.Command{
		Use:   "dep-scan",
		Short: "Scan dependencies in the local directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(scanTools) == 0 {
				scanTools = append(scanTools, "trivy")
			}
			pwd, _ := os.Getwd()
			var errs []error
			for _, tool := range scanTools {
				switch tool {
				case "trivy":
					c := tools.CreateCommand(&trivy.DirTool{
						Dir: pwd,
					})
					errs = append(errs, c.RunE(c, nil))
				default:
					return fmt.Errorf("invalid tool: %q", tool)
				}
			}
			for _, err := range errs {
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	opts.Register(c)
	flags := c.Flags()
	flags.StringSliceVarP(&scanTools, "tools", "t", nil, "Scanning tool. Defaults to \"trivy\"")
	return c
}
