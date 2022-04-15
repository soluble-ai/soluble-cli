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

package cfnpythonlint

import (
	"fmt"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
	Templates []string
}

func (t *Tool) Name() string {
	return "cfn-python-lint"
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(cmd)
	cmd.Flags().StringSliceVar(&t.Templates, "template", nil, "Explicitly specific templates in the form `t1,t2,...`.  May be repeated.  Templates must be relative to --directory.")
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "cfn-python-lint",
		Short: "Scan cloudformation templates with cfn-python-lint",
	}
}

func (t *Tool) Run() (*tools.Result, error) {
	files, err := t.findCloudformationFiles()
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no cloudformation templates found")
	}
	exec, err := t.RunDocker(&tools.DockerTool{
		Name:                "cfn-python-lint",
		DefaultNoDockerName: "cfn-lint",
		Image:               "gcr.io/soluble-repo/soluble-cfn-lint:latest",
		Directory:           t.GetDirectory(),
		Args:                append([]string{"-f", "json"}, files...),
	})
	if err != nil {
		return nil, err
	}
	result := exec.ToResult(t.GetDirectory())
	results, ok := exec.ParseJSON()
	if !ok {
		return result, nil
	}
	parseResults(result, results)
	return result, nil
}

func parseResults(result *tools.Result, results *jnode.Node) *tools.Result {
	findings := assessments.Findings{}
	for _, r := range results.Elements() {
		findings = append(findings, &assessments.Finding{
			FilePath: r.Path("Filename").AsText(),
			Line:     r.Path("Location").Path("Start").Path("LineNumber").AsInt(),
			Tool: map[string]string{
				"Level":   r.Path("Level").AsText(),
				"Message": util.TruncateRight(r.Path("Message").AsText(), 100),
				"Rule_Id": r.Path("Rule").Path("Id").AsText(),
			},
		})
	}
	result.Data = results
	result.Findings = findings
	return result
}

func (t *Tool) findCloudformationFiles() ([]string, error) {
	if len(t.Templates) > 0 {
		return t.GetFilesInDirectory(t.Templates)
	}
	return t.GetInventory().CloudformationFiles.Values(), nil
}
