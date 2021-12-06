package tfscore

import (
	"os"
	"os/exec"

	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

type PlanTool struct {
	tools.DirectoryBasedToolOpts
	TerraformPlan string
	Plan          string
	JSONPlan      string
	extraArgs     tools.ExtraArgs
}

var _ tools.Interface = &Tool{}

func (t *PlanTool) Name() string {
	return "tfscore-plan"
}

func (t *PlanTool) Register(cmd *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVar(&t.Plan, "plan", "", "Save the JSON plan including textual output to `file`")
	flags.StringVar(&t.JSONPlan, "json-plan", "", "Save the plain terraform JSON plan to `file`")
	flags.StringVar(&t.TerraformPlan, "tf-plan", "", "Save the terraform-format plan to `file`")
}

func (t *PlanTool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate a terraform plan",
		Args:  t.extraArgs.ArgsValue(),
	}
}

func (t *PlanTool) Run() (*tools.Result, error) {
	d, err := t.InstallTool(&download.Spec{Name: "tfscore"})
	if err != nil {
		return nil, err
	}
	args := []string{"plan", "-d", t.GetDirectory()}
	if t.Plan != "" {
		args = append(args, "--save-plan", t.Plan)
	}
	if t.JSONPlan != "" {
		args = append(args, "--save-json-plan", t.JSONPlan)
	}
	if t.TerraformPlan != "" {
		args = append(args, "--save-tfplan", t.TerraformPlan)
	}
	args = append(args, t.extraArgs...)
	// #nosec G204
	c := exec.Command(d.GetExePath("tfscore"), args...)
	c.Stderr = os.Stderr
	c.Stdout = os.Stderr
	t.LogCommand(c)
	return nil, c.Run()
}
