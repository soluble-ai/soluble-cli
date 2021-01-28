package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

type RunOpts struct {
	options.PrintClientOpts
	ToolVersion     string
	ToolPath        string
	SkipDockerPull  bool
	ExtraDockerArgs []string
	Internal        bool
}

var _ options.Interface = &RunOpts{}

func (o *RunOpts) Register(cmd *cobra.Command) {
	o.PrintClientOpts.Register(cmd)
	flags := cmd.Flags()
	if !o.Internal {
		flags.BoolVar(&o.SkipDockerPull, "skip-docker-pull", false, "Don't pull docker images before running them")
		flags.StringSliceVar(&o.ExtraDockerArgs, "extra-docker-args", nil, "Add extra args to invocations of docker")
		flags.StringVar(&o.ToolPath, "tool-path", "", "Run `tool` directly instead of using a CLI-managed version")
		flags.StringVar(&o.ToolVersion, "tool-version", "", "Override version of the tool to run (the image or github release name.)")
	}
}

func (o *RunOpts) RunDocker(d *DockerTool) ([]byte, error) {
	if o.ToolPath != "" {
		// don't use docker, just run it directly
		// #nosec G204
		c := exec.Command(o.ToolPath, d.Args...)
		c.Dir = d.Directory
		c.Stderr = os.Stderr
		log.Infof("Running {primary:%s} {secondary:(in %s)}", strings.Join(c.Args, " "), c.Dir)
		return c.Output()
	}
	n := o.getToolVersion(d.Name)
	if image := n.Path("image"); !image.IsMissing() {
		d.Image = image.AsText()
	}
	d.DockerArgs = append(d.DockerArgs, o.ExtraDockerArgs...)
	return d.run(o.SkipDockerPull)
}

func (o *RunOpts) InstallTool(spec *download.Spec) (*download.Download, error) {
	if o.ToolPath != "" {
		return &download.Download{
			OverrideExe: o.ToolPath,
		}, nil
	}
	if strings.HasPrefix(spec.URL, "github.com/") {
		slash := strings.LastIndex(spec.URL, "/")
		n := o.getToolVersion(spec.URL[slash+1:])
		if v := n.Path("version"); !v.IsMissing() {
			spec.RequestedVersion = v.AsText()
		}
	}
	m := download.NewManager()
	return m.Install(spec)
}

func (o *RunOpts) getToolVersion(name string) *jnode.Node {
	if o.ToolVersion != "" {
		return jnode.NewObjectNode().
			Put("image", o.ToolVersion).
			Put("version", o.ToolVersion)
	}
	temp := log.SetTempLevel(log.Error - 1)
	defer temp.Restore()
	n, err := o.GetUnauthenticatedAPIClient().Get(fmt.Sprintf("cli/tools/%s/config", name))
	if err != nil {
		return jnode.MissingNode
	}
	return n
}
