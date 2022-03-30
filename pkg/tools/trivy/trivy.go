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

package trivy

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/hashicorp/go-version"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.AssessmentOpts
	Image      string
	ClearCache bool
}

var _ tools.Single = &Tool{}

func (t *Tool) Name() string {
	return "trivy"
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.AssessmentOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVarP(&t.Image, "image", "i", "", "The image to scan")
	flags.BoolVarP(&t.ClearCache, "clear-cache", "c", false, "clear image caches and then start scanning")
	_ = cmd.MarkFlagRequired("image")
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "image-scan",
		Short: "Scan a container image for vulnerabilities of OS packages",
		Args:  cobra.ArbitraryArgs,
	}
}

func (t *Tool) Run() (*tools.Result, error) {
	d, err := t.InstallTool(&download.Spec{
		URL: "github.com/aquasecurity/trivy",
	})
	if err != nil {
		return nil, err
	}
	outfile, err := tools.TempFile("trivy*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(outfile)
	program := d.GetExePath("trivy")
	if t.ClearCache {
		err := t.runCommand(program, "image", "--clear-cache")
		if err != nil {
			return nil, err
		}
	}

	// Generate params for the scanner
	args := []string{"image", "--format", "json", "--output", outfile}
	// specify the image to scan at the end of params
	args = append(args, t.Image)

	err = t.runCommand(program, args...)
	if err != nil {
		return nil, err
	}

	dat, err := ioutil.ReadFile(outfile)
	if err != nil {
		return nil, err
	}
	n, err := jnode.FromJSON(dat)
	if err != nil {
		return nil, err
	}
	findings := assessments.Findings{}
	for _, v := range n.Path("Vulnerabilities").Elements() {
		findings = append(findings, &assessments.Finding{
			Title: v.Path("Title").AsText(),
		})
	}
	return &tools.Result{
		Data: getData(d.Version, n),
		Values: map[string]string{
			"TRIVY_VERSION": d.Version,
			"IMAGE":         t.Image,
		},
		Findings: findings,
	}, nil
}

func getData(ver string, n *jnode.Node) *jnode.Node {
	// trivy changed it's JSON format in v0.20.0
	// see https://github.com/aquasecurity/trivy/discussions/1050
	v0_20 := version.Must(version.NewVersion("0.20.0"))
	v, err := version.NewSemver(ver)
	if err == nil && v.GreaterThanOrEqual(v0_20) {
		return n.Path("Results").Get(0)
	}
	return n.Get(0)
}

func (t *Tool) runCommand(program string, args ...string) error {
	scan := exec.Command(program, args...)
	t.LogCommand(scan)
	scan.Stderr = os.Stderr
	scan.Stdout = os.Stdout
	err := scan.Run()
	if err != nil {
		return err
	}
	return nil
}
