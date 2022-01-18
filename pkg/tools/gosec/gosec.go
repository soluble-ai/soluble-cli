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

package gosec

import (
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
	return "gosec"
}

func (t *Tool) Run() (*tools.Result, error) {
	d, err := t.InstallTool(&download.Spec{
		URL: "github.com/securego/gosec",
	})
	if err != nil {
		return nil, err
	}
	args := []string{"-fmt=json", "./..."}
	// #nosec G204
	c := exec.Command(d.GetExePath("gosec"), args...)
	c.Stderr = os.Stderr
	t.LogCommand(c)
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
	result.AddValue("GOSEC_VERSION", d.Version)
	return result, nil
}

func (t *Tool) parseResults(results *jnode.Node) *tools.Result {
	findings := assessments.Findings{}
	for _, data := range results.Path("Issues").Elements() {
		file := data.Path("file").AsText()
		if t.IsExcluded(file) {
			continue
		}
		findings = append(findings, &assessments.Finding{
			FilePath: file,
			Line:     data.Path("line").AsInt(),
			Tool: map[string]string{
				"rule_id":  data.Path("rule_id").AsText(),
				"message":  data.Path("details").AsText(),
				"severity": data.Path("severity").AsText(),
				"file":     data.Path("file").AsText(),
				"line":     data.Path("line").AsText(),
				"cwe_url":  data.Path("cwe").Path("URL").AsText(),
			},
		})
	}
	resultsArray := util.RemoveJNodeElementsIf(results.Path("Issues"), func(n *jnode.Node) bool {
		return t.IsExcluded(n.Path("file").AsText())
	})
	results.Put("Issues", resultsArray)
	result := &tools.Result{
		Directory: t.GetDirectory(),
		Data:      results,
		Findings:  findings,
	}
	return result
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "gosec",
		Short: "Run gosec identify security flaws in golang",
	}
}
