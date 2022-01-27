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

package semgrep

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
	Pattern string
	Lang    string
	Config  string

	extraArgs []string
}

var _ tools.Single = &Tool{}

func (*Tool) Name() string {
	return "semgrep"
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVarP(&t.Pattern, "pattern", "e", "", "Code search pattern.")
	flags.StringVarP(&t.Lang, "lang", "l", "", "Parse pattern and all files in specified language. Must be used with -e/--pattern.")
	flags.StringVarP(&t.Config, "config", "f", "", "YAML configuration file, directory of YAML files ending in .yml|.yaml, URL of a configuration file, or semgrep registry entry name.")
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "semgrep",
		Short: "Run semgrep",
		Long:  "Run semgrep against a directory.  Any additional arguments will be passed onwards.",
		Example: `# get help
... code-scan -- --help

# look for append literal string to a string buffer (silly example)
... code-scan -e '$SB.append("...")' -l java`,
		Args: func(cmd *cobra.Command, args []string) error {
			t.extraArgs = args
			return nil
		},
	}
}

func (t *Tool) Run() (*tools.Result, error) {
	args := []string{"--json"}
	if t.Pattern != "" {
		args = append(args, "-e", t.Pattern)
	}
	if t.Lang != "" {
		args = append(args, "-l", t.Lang)
	}
	if t.Config != "" {
		args = append(args, "-f", t.Config)
	}
	customPoliciesDir, err := t.GetCustomPoliciesDir()
	if err != nil {
		return nil, err
	}
	if customPoliciesDir != "" {
		args = append(args, "-f", customPoliciesDir)
	}
	args = append(args, t.extraArgs...)
	args = append(args, ".")
	d, err := t.RunDocker(&tools.DockerTool{
		Name:            "semgrep",
		Image:           "returntocorp/semgrep:latest",
		Directory:       t.GetDirectory(),
		PolicyDirectory: customPoliciesDir,
		Args:            args,
	})
	if err != nil && (tools.IsDockerError(err) || util.ExitCode(err) != 1) {
		// semgrep exits 1 if it finds issues
		return nil, err
	}
	n, err := jnode.FromJSON(d)
	if err != nil {
		fmt.Println(string(d))
		if util.StringSliceContains(t.extraArgs, "--help") {
			return nil, nil
		}
		return nil, fmt.Errorf("could not parse JSON: %w", err)
	}
	result := t.parseResults(n)
	return result, nil
}

func (t *Tool) parseResults(n *jnode.Node) *tools.Result {
	results := n.Path("results")
	if results.Size() > 0 {
		n.Put("results", util.RemoveJNodeElementsIf(results, func(e *jnode.Node) bool {
			return t.IsExcluded(e.Path("path").AsText())
		}))
	}
	findings := assessments.Findings{}
	for _, r := range n.Path("results").Elements() {
		findings = append(findings, &assessments.Finding{
			FilePath: r.Path("path").AsText(),
			Line:     r.Path("start").Path("line").AsInt(),
			Tool: map[string]string{
				"check_id": r.Path("check_id").AsText(),
				"message":  r.Path("extra").Path("message").AsText(),
				"severity": r.Path("extra").Path("severity").AsText(),
			},
		})
	}
	result := &tools.Result{
		Data:     n,
		Findings: findings,
	}
	return result
}
