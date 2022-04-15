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

package bandit

import (
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Single = (*Tool)(nil)

func (t *Tool) Name() string {
	return "bandit"
}

func (t *Tool) Run() (*tools.Result, error) {
	args := []string{
		"--exit-zero", "-f", "json", "-r", ".",
	}
	exec, err := t.RunDocker(&tools.DockerTool{
		Name:      "bandit",
		Image:     "gcr.io/soluble-repo/soluble-bandit:latest",
		Directory: t.GetDirectory(),
		Args:      args,
	})
	if err != nil {
		return nil, err
	}
	result := exec.ToResult(t.GetDirectory())
	if !exec.ExpectExitCode(0) {
		return result, nil
	}
	n, ok := exec.ParseJSON()
	if !ok {
		return result, nil
	}
	t.parseResults(result, n)
	return result, nil
}

func (t *Tool) parseResults(result *tools.Result, results *jnode.Node) {
	findings := assessments.Findings{}
	for _, r := range results.Path("results").Elements() {
		file := r.Path("filename").AsText()
		if t.IsExcluded(file) {
			continue
		}
		findings = append(findings, &assessments.Finding{
			FilePath: file,
			Line:     r.Path("line_number").AsInt(),
			Tool: map[string]string{
				"severity":   r.Path("issue_severity").AsText(),
				"confidence": r.Path("issue_confidence").AsText(),
				"id":         r.Path("test_id").AsText(),
				"name":       r.Path("test_name").AsText(),
			},
		})
	}
	resultsArray := util.RemoveJNodeElementsIf(results.Path("results"), func(n *jnode.Node) bool {
		return t.IsExcluded(n.Path("filename").AsText())
	})
	results.Put("results", resultsArray)
	result.Data = results
	result.Findings = findings
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "bandit",
		Short: "Run bandit to identify security flaws in python",
	}
}
