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

package hadolint

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

func (t *Tool) Name() string { return "hadolint" }

func (t *Tool) Run() (*tools.Result, error) {
	// This might be a problem if we have multiple dockerfiles and they have extensions like Dockerfile.xyz
	dockerFilePath := "./Dockerfile"
	args := []string{"hadolint", "-f", "json", "-", dockerFilePath}
	exec, err := t.RunDocker(&tools.DockerTool{
		Name:                "hadolint",
		Image:               "ghcr.io/hadolint/hadolint:latest",
		DefaultNoDockerName: "hadolint",
		Directory:           t.GetDirectory(),
		Args:                args,
	})
	if err != nil {
		return nil, err
	}
	result := exec.ToResult(t.GetDirectory())
	if !exec.ExpectExitCode(0) {
		return result, nil
	}
	results, ok := exec.ParseJSON()
	if !ok {
		return result, nil
	}
	t.parseResults(result, results)
	return result, nil
}

func (t *Tool) parseResults(result *tools.Result, results *jnode.Node) {
	findings := assessments.Findings{}
	for _, data := range results.Elements() {
		file := data.Path("file").AsText()
		if t.IsExcluded(file) {
			continue
		}
		findings = append(findings, &assessments.Finding{
			FilePath: file,
			Line:     data.Path("line").AsInt(),
			Tool: map[string]string{
				"rule_id":  data.Path("code").AsText(),
				"message":  data.Path("message").AsText(),
				"severity": data.Path("level").AsText(),
				"file":     data.Path("file").AsText(),
				"line":     data.Path("line").AsText(),
			},
		})
	}
	resultsArray := util.RemoveJNodeElementsIf(results, func(n *jnode.Node) bool {
		return t.IsExcluded(n.Path("file").AsText())
	})
	results = resultsArray
	result.Data = results
	result.Findings = findings
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "hadolint",
		Short: "Run hadolint to lint your Dockerfile",
	}
}
