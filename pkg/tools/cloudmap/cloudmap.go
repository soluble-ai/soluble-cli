package cloudmap

import (
	"os"
	"os/exec"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
	StateFile string

	extraArgs tools.ExtraArgs
}

var _ tools.Interface = (*Tool)(nil)

func (*Tool) Name() string {
	return "cloud-map"
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:     "cloud-map",
		Short:   "Map cloud resources to terraform source code locations",
		Example: "Any extra arguments after -- are passed to tfscore",
		Args:    t.extraArgs.ArgsValue(),
	}
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(cmd)
	cmd.Flags().StringVar(&t.StateFile, "state-file", "", "Map resources from terraform state `file`")
}

func (t *Tool) Run() (*tools.Result, error) {
	d, err := t.InstallTool(&download.Spec{Name: "tfscore"})
	if err != nil {
		return nil, err
	}
	args := []string{"cloud-map", "-d", t.GetDirectory()}
	if t.StateFile != "" {
		args = append(args, "--state-file", t.StateFile)
	}
	args = append(args, t.extraArgs...)
	// #nosec G204
	c := exec.Command(d.GetExePath("tfscore"), args...)
	c.Stderr = os.Stderr
	t.LogCommand(c)
	dat, err := c.Output()
	if err != nil {
		return nil, err
	}
	n, err := jnode.FromJSON(dat)
	if err != nil {
		return nil, err
	}
	result := &tools.Result{
		Data:         n,
		PrintPath:    []string{"managed_resources"},
		PrintColumns: []string{"source_location.file", "source_location.line", "cloud_id"},
	}
	result.AddValue("TFSCORE_VERSION", d.Version)
	return result, nil
}
