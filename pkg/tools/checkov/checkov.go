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
	"github.com/soluble-ai/soluble-cli/pkg/redaction"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
	Framework            string
	EnableModuleDownload bool
	VarFiles             []string

	relativeVarFiles []string
	extraArgs        tools.ExtraArgs
}

var _ tools.Single = &Tool{}

var noCodeBlockChecks = map[string]bool{
	// These checkov checks will send code blocks with secrets in
	// them.  We don't want to do that.
	"CKV_AWS_41":      true,
	"CKV_BCW_1":       true,
	"CKV_LIN_1":       true,
	"CKV_OCI_1":       true,
	"CKV_OPENSTACK_1": true,
	"CKV_PAN_1":       true,
	"CKV_AWS_46":      true,
	"CKV_AWS_45":      true,
	"CKV_AZURE_45":    true,
}

func (t *Tool) Name() string {
	return "checkov"
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(cmd)
	iacbot := os.Getenv("ZODIAC_JOB_NAME") != ""
	flags := cmd.Flags()
	flags.BoolVar(&t.EnableModuleDownload, "enable-module-download", !iacbot,
		"Enable module download.  Use --enable-module-download=false to disable.")
	flags.StringSliceVar(&t.VarFiles, "var-file", nil, "Pass additional variable `files` to checkov")
}

func (t *Tool) Validate() error {
	if err := t.DirectoryBasedToolOpts.Validate(); err != nil {
		return err
	}
	if t.Framework == "" || t.Framework == "terraform" {
		for _, name := range t.VarFiles {
			if !strings.Contains(name, ".tfvars") {
				// This bug is present in at least 2.0.1021.  Strange but true.
				return fmt.Errorf("checkov only supports variable files with names that contain \".tfvars\"")
			}
			abs := name
			if !filepath.IsAbs(name) {
				abs, _ = filepath.Abs(name)
			}
			r, err := filepath.Rel(t.GetDirectory(), abs)
			if err != nil || strings.ContainsRune(r, os.PathSeparator) {
				// This is also a limitation of checkov
				return fmt.Errorf("variable files must be in the same directory as the target terraform")
			}
			t.relativeVarFiles = append(t.relativeVarFiles, r)
		}
	}
	return nil
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "checkov",
		Short: "Scan terraform for security vulnerabilities",
		Args:  t.extraArgs.ArgsValue(),
	}
}

func (t *Tool) Run() (*tools.Result, error) {
	dt := &tools.DockerTool{
		Name:                "checkov",
		Image:               "bridgecrew/checkov:latest",
		DefaultNoDockerName: "checkov",
		Args: []string{
			"-o", "json", "-s",
		},
	}
	if t.RepoRoot != "" {
		// We want to run in the repo root and target a relative directory under
		// that so the module references to peer or sibling directories
		// resolve correctly.
		dt.Directory = t.RepoRoot
		targetDir, _ := filepath.Rel(t.RepoRoot, t.GetDirectory())
		dt.WorkingDirectory = targetDir
	} else {
		dt.Directory = t.GetDirectory()
	}
	if t.Framework == "helm" {
		// for Helm we use the -f option, this avoids checkov scanning
		// itself for Chart.yaml files and avoiding subcharts
		if !util.FileExists(fmt.Sprintf("%s/Chart.yaml", t.GetDirectory())) {
			return nil, fmt.Errorf("%s does not contain Chart.yaml", t.GetDirectory())
		}
		dt.AppendArgs("-f", "./Chart.yaml")
	} else {
		dt.AppendArgs("-d", ".")
	}
	if t.Framework != "" {
		dt.AppendArgs("--framework", t.Framework)
	}
	if t.Framework == "terraform" && t.EnableModuleDownload {
		dt.AppendArgs("--download-external-modules", "true")
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
		dt.AppendArgs("--external-checks-dir", customPoliciesDir)
		dt.Mount(customPoliciesDir, "/policy")
	}
	for _, varFile := range t.relativeVarFiles {
		dt.AppendArgs("--var-file", varFile)
	}
	dt.AppendArgs(t.extraArgs...)
	if t.Framework == "" || t.Framework == "terraform" {
		propagateTfVarsEnv(dt, os.Environ())
	}
	exec, err := t.RunDocker(dt)
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
	t.processResults(result, n)
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

func (t *Tool) processResults(result *tools.Result, data *jnode.Node) *tools.Result {
	result.Data = data
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
		if codeBlockIsSensitive(n, pass) {
			n.Remove("code_block")
		}
		filePath := n.Path("file_path").AsText()
		if len(filePath) > 0 && filePath[0] == '/' {
			// checkov sticks an extra slash at the beginning
			filePath = filePath[1:]
		}
		if t.Framework == "helm" {
			// checkov generates templates with helm, so the "file_path" doesn't
			// actually match the path in the repo.  We'll rewrite so it does.
			// Two things:
			// 1 - filePath may a tmp file in /tmp
			// 2 - filePath will start with the name of the chart
			// directory
			base := filepath.Base(t.GetDirectory())
			parts := strings.Split(filepath.ToSlash(filePath), "/")
			if len(parts) > 3 && parts[0] == "tmp" && strings.HasPrefix(parts[1], "tmp") {
				if parts[2] == base {
					filePath = strings.Join(parts[3:], "/")
				} else {
					filePath = strings.Join(parts[2:], "/")
				}
			} else if len(parts) > 1 && parts[0] == base {
				filePath = strings.Join(parts[1:], "/")
			}
		}
		n.Put("file_path", filePath)
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

func propagateTfVarsEnv(d *tools.DockerTool, env []string) {
	// checkov (as of 2.0.1021) has support for TF_VAR_ environment
	// variables but it does not work properly.  We'll pass them
	// along so that when checkov fixes the issue it will work.
	for _, ev := range env {
		eq := strings.IndexRune(ev, '=')
		name := ev[0:eq]
		if strings.HasPrefix(name, "TF_VAR_") {
			d.PropagateEnvironmentVars = append(d.PropagateEnvironmentVars, name)
		}
	}
}

func codeBlockIsSensitive(check *jnode.Node, pass bool) bool {
	if !pass {
		checkID := check.Path("check_id").AsText()
		if noCodeBlockChecks[checkID] {
			return true
		}
	}
	for _, elt := range check.Path("code_block").Elements() {
		text := elt.Get(1).AsText()
		if redaction.ContainsSecret(text) {
			return true
		}
	}
	return false
}
