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

package tfscore

import (
	"os"
	"os/exec"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
	Region           string
	TerraformVersion string
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "tfscore"
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVarP(&t.Region, "aws-region", "r", "", "AWS Region where resources exit") // this needs to be removed when tfscore reads from plan output
}

func (t *Tool) Run() (*tools.Result, error) {
	args := []string{"plan"}

	if t.Region != "" {
		args = append(args, "--aws.region", t.Region)
	}
	args = append(args, "-d", t.GetDirectory())

	d, err := t.InstallTool(&download.Spec{
		URL:  "https://storage.googleapis.com/storage/v1/b/soluble-public/o/tfscore%2Fv0.1.9%2Ftfscore_0.1.9_darwin_amd64.tar.gz?alt=media",
		Name: "tfscore",
	})
	if err != nil {
		return nil, err
	}

	// run tfscore
	// #nosec G204
	c := exec.Command(d.GetExePath("tfscore"), args...)
	c.Stderr = os.Stderr
	log.Infof("Running {primary:%s}", strings.Join(c.Args, " "))
	output, err := c.Output()
	if util.ExitCode(err) == 1 {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	n, err := jnode.FromJSON(output)
	if err != nil {
		return nil, err
	}

	result := t.parseResults(n)
	result.AddValue("TFSCORE_VERSION", d.Version)
	return result, nil
}

func (t *Tool) parseResults(n *jnode.Node) *tools.Result {
	return &tools.Result{
		Directory: t.GetDirectory(),
		Data:      n,
	}
}
