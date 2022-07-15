package tfplanscan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {

	plan := &checkov.Plan{}
	c := tools.CreateCommand(plan)
	c.Use = "terraform-plan-scan"
	c.Short = "Scan a terraform plan"
	c.Long = `Scan a terraform plan

Use the sub-commands to explicitly choose a scanner to use.`
	ckv := tools.CreateCommand(&checkov.Plan{})
	ckv.Use = "checkov"
	ckv.Short = "Scan a terraform plan with checkov"
	c.AddCommand(ckv)
	return c
}
