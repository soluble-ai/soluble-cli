// Copyright 2021 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type RunOpts struct {
	options.PrintClientOpts
	ToolVersion     string
	ToolPath        string
	SkipDockerPull  bool
	ExtraDockerArgs []string
	NoDocker        bool
	Internal        bool
	Quiet           bool
}

var _ options.Interface = &RunOpts{}

func (o *RunOpts) GetRunHiddenOptions() *options.HiddenOptionsGroup {
	return &options.HiddenOptionsGroup{
		Name: "run-options",
		Long: "Options for running tools",
		CreateFlagsFunc: func(flags *pflag.FlagSet) {
			flags.BoolVar(&o.SkipDockerPull, "skip-docker-pull", false, "Don't pull docker images before running them")
			flags.StringSliceVar(&o.ExtraDockerArgs, "extra-docker-args", nil, "Add extra args to invocations of docker")
			flags.StringVar(&o.ToolPath, "tool-path", "", "Run `tool` directly instead of using a CLI-managed version")
			flags.StringVar(&o.ToolVersion, "tool-version", "", "Override version of the tool to run (the image or github release name.)")
			flags.BoolVar(&o.NoDocker, "no-docker", false, "Always run tools locally instead of using Docker")
		},
	}
}

func (o *RunOpts) Register(cmd *cobra.Command) {
	o.PrintClientOpts.Register(cmd)
	if !o.Internal {
		o.GetRunHiddenOptions().Register(cmd)
	}
}

func (o *RunOpts) UsingDocker() bool {
	return o.ToolPath == "" && !o.NoDocker
}

// Run a docker tool.  If the tool cannot be run because docker isn't running or
// the tool path isn't known then returns an error.  Otherwise returns an ExecuteResult
// that holds the output, log and exit code of the command.
func (o *RunOpts) RunDocker(d *DockerTool) (*ExecuteResult, error) {
	if !o.UsingDocker() {
		path := o.ToolPath
		if path == "" {
			path = d.DefaultNoDockerName
		}
		if path == "" {
			path = d.Name
		}
		if path == "" {
			return nil, fmt.Errorf("cannot run this tool locally, use --tool-path to explicitly name the local program")
		}
		// don't use docker, just run it directly
		// #nosec G204
		c := exec.Command(path, d.Args...)
		c.Dir = d.Directory
		c.Stderr = os.Stderr
		return o.ExecuteCommand(c), nil
	}
	n := o.getToolVersion(d.Name)
	if image := n.Path("image"); !image.IsMissing() {
		d.Image = image.AsText()
	}
	d.DockerArgs = append(d.DockerArgs, o.ExtraDockerArgs...)
	d.Quiet = o.Quiet
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

func (o *RunOpts) InstallAPIServerArtifact(name, urlPath string) (*download.Download, error) {
	apiClient := o.GetAPIClient()
	m := download.NewManager()
	return m.Install(&download.Spec{
		Name:                       name,
		APIServerArtifact:          urlPath,
		APIServer:                  apiClient,
		LatestReleaseCacheDuration: 1 * time.Minute,
	})
}

func (o *RunOpts) ExecuteCommand(c *exec.Cmd) *ExecuteResult {
	o.LogCommand(c)
	return executeCommand(c)
}

func (o *RunOpts) LogCommand(c *exec.Cmd) {
	if o.Quiet {
		return
	}
	if c.Dir != "" {
		log.Infof("Running {primary:%s} {secondary:(in %s)}", strings.Join(c.Args, " "), c.Dir)
		return
	}
	log.Infof("Running {primary:%s}", strings.Join(c.Args, " "))
}
