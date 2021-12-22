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

package yarnaudit

import (
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Single = (*Tool)(nil)

func (t *Tool) Name() string {
	return "yarn-audit"
}

func (t *Tool) Run() (*tools.Result, error) {
	args := []string{"audit", "-s", "--json"}
	d, _ := t.RunDocker(&tools.DockerTool{
		Name:                "yarn-audit",
		Image:               "gcr.io/soluble-repo/soluble-yarn:latest",
		DefaultNoDockerName: "yarn",
		Directory:           t.GetDirectory(),
		Args:                args,
	})
	results, err := jnode.FromJSON(d)
	if err != nil {
		if d != nil {
			os.Stderr.Write(d)
		}
		return nil, err
	}
	result := t.parseResults(results)
	return result, nil
}

func (t *Tool) parseResults(results *jnode.Node) *tools.Result {
	findings := assessments.Findings{}
	for _, data := range results.Elements() {
		if data.Path("type").AsText() == "auditAdvisory" {
			findings = append(findings, &assessments.Finding{
				Tool: map[string]string{
					"id":             data.Path("advisory").Path("id").AsText(),
					"cwe":            data.Path("advisory").Path("cwe").AsText(),
					"module":         data.Path("advisory").Path("module_name").AsText(),
					"recommendation": data.Path("advisory").Path("recommendation").AsText(),
					"severity":       data.Path("advisory").Path("severity").AsText(),
					"title":          data.Path("advisory").Path("title").AsText(),
				},
			})
		}
	}
	result := &tools.Result{
		Directory: t.GetDirectory(),
		Data:      results,
		Findings:  findings,
		PrintColumns: []string{
			"tool.id", "tool.title", "tool.recommendation", "tool.severity", "tool.module", "tool.cwe",
		},
	}
	return result
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "yarn-audit",
		Short: "Run yarn audit to find vulnerable dependencies of a yarn application",
	}
}
