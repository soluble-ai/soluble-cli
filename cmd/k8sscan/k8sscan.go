package k8sscan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&checkov.Tool{
		Framework: "kubernetes",
	})
	c.Use = "kubernetes-scan"
	c.Short = "Scan kubernetes manifests with checkov"
	c.Aliases = []string{"k8s-scan"}
	return c
}
