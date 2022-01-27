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

package brakeman

import (
	"os"

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
	return "brakeman"
}

func (t *Tool) Run() (*tools.Result, error) {
	args := []string{"-f", "json", "-q"}
	d, err := t.RunDocker(&tools.DockerTool{
		Name:                "brakeman",
		Image:               "gcr.io/soluble-repo/soluble-brakeman:latest",
		DefaultNoDockerName: "brakeman",
		Directory:           t.GetDirectory(),
		Args:                args,
	})
	if err != nil && tools.IsDockerError(err) {
		return nil, err
	}
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
	for _, data := range results.Path("warnings").Elements() {
		file := data.Path("file").AsText()
		if t.IsExcluded(file) {
			continue
		}
		findings = append(findings, &assessments.Finding{
			FilePath: file,
			Line:     data.Path("line").AsInt(),
			Tool: map[string]string{
				"type":       data.Path("warning_type").AsText(),
				"code":       data.Path("warning_code").AsText(),
				"message":    data.Path("message").AsText(),
				"file":       data.Path("file").AsText(),
				"confidence": data.Path("confidence").AsText(),
				"method":     data.Path("location").Path("method").AsText(),
				"line":       data.Path("line").AsText(),
			},
		})
	}
	resultsArray := util.RemoveJNodeElementsIf(results.Path("warnings"), func(n *jnode.Node) bool {
		return t.IsExcluded(n.Path("file").AsText())
	})
	results.Put("warnings", resultsArray)
	result := &tools.Result{
		Directory: t.GetDirectory(),
		Data:      results,
		Findings:  findings,
	}
	return result
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "brakeman",
		Short: "Run static analysis to find security vulnerabilities in Ruby on Rails applications",
	}
}
