package tfscore

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
	PlanFile   string
	SaveBundle string

	extraArgs tools.ExtraArgs
}

var _ tools.Single = &Tool{}

func (t *Tool) Name() string {
	return "tfscore"
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVar(&t.PlanFile, "plan", "", "Scan the terraform plan in `file`")
	flags.StringVar(&t.SaveBundle, "save-bundle", "",
		"Write a development bundle tar file to `file`")
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan a terraform plan",
		Args:  t.extraArgs.ArgsValue(),
	}
}

func (t *Tool) Validate() error {
	if t.PlanFile == "" && t.SaveBundle == "" {
		return fmt.Errorf("--plan or --save-bundle is required")
	}
	if t.Directory == "" {
		t.Directory = filepath.Dir(t.PlanFile)
	}
	return t.DirectoryBasedToolOpts.Validate()
}

func (t *Tool) Run() (*tools.Result, error) {
	d, err := t.InstallTool(&download.Spec{Name: "tfscore"})
	if err != nil {
		return nil, err
	}
	var scorePath string
	args := []string{"score", "-d", t.GetDirectory()}
	if t.PlanFile != "" {
		scorePath, err = util.TempFile("tfscore")
		if err != nil {
			return nil, err
		}
		args = append(args, "--plan", t.PlanFile, "--save-score", scorePath)
	}
	if t.SaveBundle != "" {
		args = append(args, "--save-bundle", t.SaveBundle)
		if t.PlanFile == "" {
			args = append(args, "--dont-score")
		}
	}
	args = append(args, t.extraArgs...)
	// #nosec G204
	c := exec.Command(d.GetExePath("tfscore"), args...)
	c.Stderr = os.Stderr
	c.Stdout = os.Stderr
	exec := t.ExecuteCommand(c)
	result := exec.ToResult(t.GetDirectory())
	if !exec.ExpectExitCode(0) {
		return result, nil
	}
	if scorePath != "" {
		dat, err := os.ReadFile(scorePath)
		if err != nil {
			exec.SetFailureFromError(tools.GarbledResultFailure, err)
			return result, nil
		}
		n, err := jnode.FromJSON(dat)
		if err != nil {
			exec.SetFailureFromError(tools.GarbledResultFailure, err)
			return result, nil
		}
		findings := assessments.Findings{}
		for _, f := range n.Path("risks").Elements() {
			findings = append(findings,
				&assessments.Finding{
					SID:         f.Path("id").AsText(),
					Severity:    f.Path("severity").AsText(),
					FilePath:    f.Path("file").AsText(),
					Line:        f.Path("line").AsInt(),
					Description: f.Path("message").AsText(),
				},
			)
		}
		result.Data = n
		result.Findings = findings
		result.AddValue("TFSCORE_VERSION", d.Version)
	}
	return result, nil
}
