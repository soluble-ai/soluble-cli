package imagescan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/trivy"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return tools.CreateCommand(&trivy.Tool{},
		&cobra.Command{
			Use:   "image-scan",
			Short: "Scan a container image",
		})
}
