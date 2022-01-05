package cdkscan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&checkov.CDK{})
	c.Use = "cdk-scan"
	c.Short = "Scan Amazon CDK infrastructure-as-code"
	return c
}
