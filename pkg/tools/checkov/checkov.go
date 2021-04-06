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

package checkov

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
	Framework string

	extraArgs []string
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "checkov"
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "checkov",
		Short: "Scan with checkov",
		Example: `# Any additional args after -- are passed through to checkov, eg:
... checkov -- --help`,
		Args: func(cmd *cobra.Command, args []string) error {
			t.extraArgs = args
			return nil
		},
	}
}

func (t *Tool) Run() (*tools.Result, error) {
	args := []string{
		"-d", ".", "-o", "json", "-s",
	}
	if t.Framework != "" {
		args = append(args, "--framework", t.Framework)
	}
	customPoliciesDir, err := t.GetCustomPoliciesDir()
	if err != nil {
		return nil, err
	}
	if customPoliciesDir != "" {
		args = append(args, "--external-checks-dir", customPoliciesDir)
	}
	args = append(args, t.extraArgs...)
	dat, err := t.RunDocker(&tools.DockerTool{
		Name:             "checkov",
		Image:            "gcr.io/soluble-repo/checkov:latest",
		DefaultLocalPath: "checkov",
		Directory:        t.GetDirectory(),
		PolicyDirectory:  customPoliciesDir,
		Args:             args,
	})
	if err != nil {
		if dat != nil {
			_, _ = os.Stderr.Write(dat)
		}
		return nil, err
	}
	n, err := jnode.FromJSON(dat)
	if err != nil {
		_, _ = os.Stderr.Write(dat)
		return nil, err
	}

	result := t.processResults(n)
	return result, nil
}

func (t *Tool) processResults(data *jnode.Node) *tools.Result {
	result := &tools.Result{
		Directory: t.GetDirectory(),
		Data:      data,
		PrintPath: []string{},
		PrintColumns: []string{
			"tool.check_id", "pass", "tool.check_type", "filePath", "line", "title",
		},
	}
	if data.IsArray() {
		// checkov returns an array if it runs more than one check type at a go
		for _, n := range data.Elements() {
			t.processCheckResults(result, n)
		}
	} else {
		t.processCheckResults(result, data)
	}
	return result
}

func (t *Tool) processCheckResults(result *tools.Result, e *jnode.Node) {
	checkType := e.Path("check_type").AsText()
	results := e.Path("results")
	passedChecks := t.processChecks(result, results.Path("passed_checks"), checkType, true)
	failedChecks := t.processChecks(result, results.Path("failed_checks"), checkType, false)
	updateChecks(results, "passed_checks", passedChecks)
	updateChecks(results, "failed_checks", failedChecks)
	result.AddValue("CHECKOV_VERSION", e.Path("summary").Path("checkov_version").AsText())
}

func updateChecks(results *jnode.Node, name string, checks *jnode.Node) {
	if checks.Size() == 0 {
		results.Remove(name)
	} else {
		results.Put(name, checks)
	}
}

func (t *Tool) processChecks(result *tools.Result, checks *jnode.Node, checkType string, pass bool) *jnode.Node {
	for _, n := range checks.Elements() {
		filePath := n.Path("file_path").AsText()
		if len(filePath) > 0 && filePath[0] == '/' {
			filePath = filePath[1:]
			n.Put("file_path", filePath)
		}
	}
	checks = util.RemoveJNodeElementsIf(checks, func(e *jnode.Node) bool {
		return t.IsExcluded(e.Path("file_path").AsText())
	})
	for _, n := range checks.Elements() {
		result.Findings = append(result.Findings, &assessments.Finding{
			Tool: map[string]string{
				"check_id":   n.Path("check_id").AsText(),
				"check_type": checkType,
			},
			FilePath: n.Path("file_path").AsText(),
			Line:     n.Path("file_line_range").Get(0).AsInt(),
			Pass:     pass,
			Title:    n.Path("check_name").AsText(),
		})
	}
	return checks
}
