package imagescan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/trivy"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&trivy.Tool{})
	c.Use = "image-scan"
	c.Short = "Scan a container image"
	return c
}
