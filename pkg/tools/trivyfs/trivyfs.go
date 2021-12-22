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

package trivyfs

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Single = &Tool{}

func (t *Tool) Name() string {
	return "trivy-fs"
}

func (t *Tool) Register(c *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(c)
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "trivy",
		Short: "Run trivy in filesystem mode to scan app dependencies",
		Long: `Scan dependencies of an app with trivy.
		
Trivy will look for vulnerabilities based on lock files such as Gemfile.lock and package-lock.json.`,
	}
}

func (t *Tool) Run() (*tools.Result, error) {
	d, err := t.InstallTool(&download.Spec{
		URL: "github.com/aquasecurity/trivy",
	})
	if err != nil {
		return nil, err
	}
	outfile, err := tools.TempFile("trivyfs*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(outfile)
	program := d.GetExePath("trivy")
	args := []string{"fs", "--format", "json", "--output", outfile, t.GetDirectory()}
	c := exec.Command(program, args...)
	c.Stderr = os.Stderr
	c.Stdout = os.Stderr
	t.LogCommand(c)
	if err := c.Run(); err != nil {
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
	n = util.RemoveJNodeElementsIf(n, func(e *jnode.Node) bool {
		return t.IsExcluded(e.Path("Target").AsText())
	})

	result := &tools.Result{
		Data: n,
		Values: map[string]string{
			"TRIVY_VERSION": d.Version,
		},
		Findings:     parseResults(n),
		PrintColumns: []string{"filePath", "tool.PkgName", "tool.VulnerabilityID", "tool.Severity", "tool.InstalledVersion", "tool.FixedVersion", "title"},
	}

	return result, nil
}

func parseResults(n *jnode.Node) []*assessments.Finding {
	findings := []*assessments.Finding{}
	for _, e := range n.Elements() {
		target := e.Path("Target").AsText()
		for _, v := range e.Path("Vulnerabilities").Elements() {
			f := &assessments.Finding{
				FilePath: target,
				Title:    v.Path("Title").AsText(),
			}
			copyAttrs(f, v, "VulnerabilityID", "PkgName", "InstalledVersion", "FixedVersion", "Severity")
			findings = append(findings, f)
		}
	}
	return findings
}

func copyAttrs(finding *assessments.Finding, v *jnode.Node, names ...string) {
	for _, name := range names {
		finding.SetAttribute(name, v.Path(name).AsText())
	}
}
