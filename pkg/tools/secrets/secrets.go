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

package secrets

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

	args []string
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "secrets"
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Args: func(cmd *cobra.Command, args []string) error {
			t.args = args
			return nil
		},
	}
}

func (t *Tool) Run() (*tools.Result, error) {
	// --all-files includes files not checked into git
	// --no-verify avoids making network calls to check credentials
	var args []string
	if t.NoDocker || t.ToolPath != "" {
		// the image entrypoint sticks scan in the args
		args = append(args, "scan")
	}
	args = append(args, "--all-files", "--no-verify")
	args = append(args, t.args...)
	customPoliciesDir, err := t.GetCustomPoliciesDir()
	if err != nil {
		return nil, err
	}
	if customPoliciesDir != "" {
		args = append(args, "--custom-plugins", customPoliciesDir)
	}
	d, _ := t.RunDocker(&tools.DockerTool{
		Name:                "soluble-secrets",
		DefaultNoDockerName: "detect-secrets",
		Image:               "gcr.io/soluble-repo/soluble-secrets:latest",
		Directory:           t.GetDirectory(),
		PolicyDirectory:     customPoliciesDir,
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

func (t *Tool) excludeResults(results *jnode.Node) {
	rs := results.Path("results")
	if rs.Size() > 0 {
		results.Put("results", util.RemoveJNodeEntriesIf(rs, func(k string, v *jnode.Node) bool {
			return t.IsExcluded(k)
		}))
	}
}

func (t *Tool) parseResults(results *jnode.Node) *tools.Result {
	t.excludeResults(results)
	findings := []*assessments.Finding{}
	for k, v := range results.Path("results").Entries() {
		if v.IsArray() {
			for _, p := range v.Elements() {
				p.Put("file_name", k)
				findings = append(findings, &assessments.Finding{
					FilePath: k,
					Line:     p.Path("line_number").AsInt(),
					Title:    p.Path("type").AsText(),
				})
			}
		}
	}
	result := &tools.Result{
		Directory:    t.GetDirectory(),
		Data:         results,
		Findings:     findings,
		PrintColumns: []string{"filePath", "line", "title"},
	}
	return result
}
