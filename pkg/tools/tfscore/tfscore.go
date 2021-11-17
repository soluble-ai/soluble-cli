package tfscore

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
	PlanFile string

	extraArgs tools.ExtraArgs
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "tfscore"
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(cmd)
	cmd.Flags().StringVar(&t.PlanFile, "plan", "", "Score the terraform plan in `file`")
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:     "plan",
		Short:   "Score a terraform plan with tfscore",
		Example: "Any extra arguments after -- are passed to tfscore",
		Args:    t.extraArgs.ArgsValue(),
	}
}

func (t *Tool) Validate() error {
	if t.PlanFile == "" {
		return fmt.Errorf("--plan is required")
	}
	if t.Directory == "" {
		t.Directory = filepath.Dir(t.PlanFile)
	}
	return t.GetDirectoryBasedToolOptions().Validate()
}

func (t *Tool) Run() (*tools.Result, error) {
	d, err := t.InstallTool(&download.Spec{Name: "tfscore"})
	if err != nil {
		return nil, err
	}
	scorePath, err := util.TempFile("tfscore")
	if err != nil {
		return nil, err
	}
	args := []string{"score", "-d", t.GetDirectory(), "--skip-init",
		"--read-plan", t.PlanFile, "--save-score", scorePath}
	args = append(args, t.extraArgs...)
	// #nosec G204
	c := exec.Command(d.GetExePath("tfscore"), args...)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	t.LogCommand(c)
	if err := c.Run(); err != nil {
		return nil, err
	}
	dat, err := os.ReadFile(scorePath)
	if err != nil {
		return nil, err
	}
	n, err := jnode.FromJSON(dat)
	if err != nil {
		return nil, err
	}
	result := &tools.Result{
		Data:         n,
		Directory:    t.GetDirectory(),
		PrintPath:    []string{"risks"},
		PrintColumns: []string{"id", "severity", "file", "line", "message"},
	}
	result.AddValue("TFSCORE_VERSION", d.Version)
	return result, nil
}
