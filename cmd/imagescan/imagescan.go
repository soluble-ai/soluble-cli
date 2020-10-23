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
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.RunTool(tool)
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVar(&tool.Image, "image", "", "The image to scan")
	_ = c.MarkFlagRequired("image")
	return c
}
