package imagescan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/trivy"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	tool := &trivy.Tool{}
	opts := &tools.ToolOpts{}
	c := &cobra.Command{
		Use:   "image-scan",
		Short: "Scan a container image and optionally upload the results",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.RunTool(tool)
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVarP(&tool.Image, "image", "i", "", "The image to scan")
	flags.BoolVarP(&tool.ClearCache, "clear-cache", "c", false, "clear image caches and then start scanning")
	flags.BoolVarP(&tool.IgnoreUnfixed, "ignore-unfixed", "u", false, "display only fixed vulnerabilities")
	_ = c.MarkFlagRequired("image")
	return c
}
