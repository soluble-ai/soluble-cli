package cloudmap

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.ToolOpts
	tools.DirectoryOpt
	tools.UploadOpts
	StateFile string

	extraArgs tools.ExtraArgs
}

var _ tools.Simple = (*Tool)(nil)

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

func (t *Tool) Validate() error {
	if err := t.DirectoryOpt.Validate(&t.ToolOpts); err != nil {
		return err
	}
	if err := t.ToolOpts.Validate(); err != nil {
		return err
	}
	return nil
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.ToolOpts.Register(cmd)
	t.DirectoryOpt.Register(cmd)
	t.UploadOpts.Register(cmd)
	cmd.Flags().StringVar(&t.StateFile, "state-file", "", "Map resources from terraform state `file`")
	t.Path = []string{"managed_resources"}
	t.Columns = []string{"source_location.file", "source_location.line", "cloud_id"}
}

func (t *Tool) Run() error {
	d, err := t.InstallTool(&download.Spec{Name: "tfscore"})
	if err != nil {
		return err
	}
	args := []string{"cloud-map", "-d", t.GetDirectory()}
	if t.StateFile != "" {
		args = append(args, "--state-file", t.StateFile)
	}
	args = append(args, t.extraArgs...)
	// #nosec G204
	c := exec.Command(d.GetExePath("tfscore"), args...)
	c.Stderr = os.Stderr
	exec := t.ExecuteCommand(c)
	if !exec.ExpectExitCode(0) {
		// Future: upload error
		return exec.ToError()
	}
	dat := exec.Output
	n, err := jnode.FromJSON(dat)
	if err != nil {
		return err
	}
	t.PrintResult(n)
	if t.UploadEnabled {
		values := t.GetStandardXCPValues()
		values["TFSCORE_VERSION"] = d.Version
		exec.SetUploadValues(values)
		options := []api.Option{
			xcp.WithCIEnv(t.GetDirectory()),
			xcp.WithFileFromReader("cloudmap", "cloudmap.json", bytes.NewReader(dat)),
		}
		options = exec.AppendUploadOptions(t.CompressFiles, options)
		_, err := t.GetAPIClient().XCPPost(t.GetOrganization(), "cloudmap", nil, values, options...)
		if err != nil {
			return err
		}
	}
	return nil
}
