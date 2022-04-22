package tfscore

import (
	"os"
	"os/exec"

	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

type PlanTool struct {
	tools.ToolOpts
	tools.DirectoryOpt
	TerraformPlan string
	Plan          string
	JSONPlan      string
	extraArgs     tools.ExtraArgs
}

var _ tools.Simple = &PlanTool{}

func (t *PlanTool) Name() string {
	return "tfscore-plan"
}

func (t *PlanTool) Register(cmd *cobra.Command) {
	t.ToolOpts.Register(cmd)
	t.DirectoryOpt.Register(cmd)
	flags := cmd.Flags()
	flags.StringVar(&t.Plan, "plan", "", "Save the JSON plan including textual output to `file`")
	flags.StringVar(&t.JSONPlan, "json-plan", "", "Save the plain terraform JSON plan to `file`")
	flags.StringVar(&t.TerraformPlan, "tf-plan", "", "Save the terraform-format plan to `file`")
}

func (t *PlanTool) Validate() error {
	if err := t.ToolOpts.Validate(); err != nil {
		return err
	}
	return t.DirectoryOpt.Validate(&t.ToolOpts)
}

func (t *PlanTool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate a terraform plan",
		Args:  t.extraArgs.ArgsValue(),
	}
}

func (t *PlanTool) Run() error {
	d, err := t.InstallTool(&download.Spec{Name: "tfscore"})
	if err != nil {
		return err
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
	exec := t.ExecuteCommand(c)
	if !exec.ExpectExitCode(0) {
		return exec.ToError()
	}
	return nil
}
