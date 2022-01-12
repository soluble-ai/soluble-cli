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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
	Framework            string
	EnableModuleDownload bool

	extraArgs tools.ExtraArgs
}

var _ tools.Single = &Tool{}

func (t *Tool) Name() string {
	return "checkov"
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(cmd)
	iacbot := os.Getenv("ZODIAC_JOB_NAME") != ""
	flags := cmd.Flags()
	flags.BoolVar(&t.EnableModuleDownload, "enable-module-download", !iacbot,
		"Enable module download.  Use --enable-module-download=false to disable.")
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "checkov",
		Short: "Scan terraform for security vulnerabilities",
		Args:  t.extraArgs.ArgsValue(),
	}
}

func (t *Tool) Run() (*tools.Result, error) {
	args := []string{
		"-o", "json", "-s",
	}
	if t.Framework != "" {
		args = append(args, "--framework", t.Framework)
	}
	if t.Framework == "terraform" && t.EnableModuleDownload {
		args = append(args, "--download-external-modules", "true")
	}
	if t.Framework == "helm" && (t.NoDocker || t.ToolPath != "") {
		if err := t.makeHelmAvailable(); err != nil {
			return nil, err
		}
	}
	customPoliciesDir, err := t.GetCustomPoliciesDir()
	if err != nil {
		return nil, err
	}
	if customPoliciesDir != "" {
		args = append(args, "--external-checks-dir", customPoliciesDir)
	}
	args = append(args, t.extraArgs...)
	toolDir := t.GetDirectory()
	if t.RepoRoot != "" {
		dir, _ := filepath.Rel(t.RepoRoot, toolDir)
		toolDir = t.RepoRoot
		args = append(args, "-d", dir)
	} else {
		args = append(args, "-d", ".")
	}
	dat, err := t.RunDocker(&tools.DockerTool{
		Name:                "checkov",
		Image:               "bridgecrew/checkov:latest",
		DefaultNoDockerName: "checkov",
		Directory:           toolDir,
		PolicyDirectory:     customPoliciesDir,
		Args:                args,
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

func (t *Tool) makeHelmAvailable() error {
	c := exec.Command("helm", "version")
	if err := c.Run(); err != nil {
		// helm is not installed, so install it from github
		installer := &tools.RunOpts{}
		d, err := installer.InstallTool(&download.Spec{
			URL: "github.com/helm/helm",
		})
		if err != nil {
			return err
		}
		dir := filepath.Dir(d.GetExePath("helm"))
		// add to path
		path := os.Getenv("PATH")
		if path == "" {
			path = dir
		} else {
			path = fmt.Sprintf("%s%c%s", path, os.PathListSeparator, dir)
		}
		log.Infof("Adding {info:%s} to PATH", dir)
		os.Setenv("PATH", path)
	}
	return nil
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
	if rc := e.Path("summary").Path("resource_count"); !rc.IsMissing() {
		result.AddValue("RESOURCE_COUNT", rc.AsText())
	}
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
		if t.Framework == "helm" {
			// checkov generates templates with helm, so the "file_path" doesn't
			// actually match the path in the repo.  We'll rewrite so it does.
			base := filepath.Base(t.GetDirectory()) + "/"
			if strings.HasPrefix(filepath.ToSlash(filePath), base) {
				filePath = filePath[len(base):]
				n.Put("file_path", filePath)
			}
		}
	}
	checks = util.RemoveJNodeElementsIf(checks, func(e *jnode.Node) bool {
		return t.IsExcluded(e.Path("file_path").AsText())
	})
	for _, n := range checks.Elements() {
		path := n.Path("file_path").AsText()
		finding := &assessments.Finding{
			Tool: map[string]string{
				"check_id":   n.Path("check_id").AsText(),
				"check_type": checkType,
			},
			FilePath:      path,
			Line:          n.Path("file_line_range").Get(0).AsInt(),
			Pass:          pass,
			Title:         n.Path("check_name").AsText(),
			GeneratedFile: t.isGeneratedFile(path),
		}
		result.Findings = append(result.Findings, finding)
	}
	return checks
}

func (t *Tool) isGeneratedFile(path string) bool {
	if t.Framework == "" || t.Framework == "terraform" {
		return strings.HasPrefix(filepath.ToSlash(path), ".external_modules/")
	}
	return false
}
