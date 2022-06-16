package kustomizescan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&checkov.Kustomize{})
	c.Use = "kustomize-scan"
	c.Short = "Scan kustomize templates"
	c.Long = `Scan Kubernetes kustomize templates.

Use the sub-commands to explicitly choose a scanner to use.`
	ckv := tools.CreateCommand(&checkov.Kustomize{})
	ckv.Short = "Scan kustomize templates with checkov"
	c.AddCommand(ckv)
	return c
}
