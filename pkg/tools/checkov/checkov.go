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
	"path/filepath"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
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

	relativeVarFiles    []string
	extraArgs           tools.ExtraArgs
	pathTranslationFunc func(string) string
	workingDir          string
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
	if t.Framework == "terraform" {
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
		Image:               "gcr.io/soluble-repo:2",
		DefaultNoDockerName: "checkov",
		Args: []string{
			"-o", "json", "-s", "--skip-download",
		},
	}
	if t.UsingDocker() && t.RepoRoot != "" {
		// We want to run in the repo root and target a relative directory under
		// that so the module references to peer or sibling directories
		// resolve correctly.
		dt.Directory = t.RepoRoot
		workingDir := t.workingDir
		if t.workingDir == "" {
			workingDir = t.GetDirectory()
		}
		targetDir, err := filepath.Rel(t.RepoRoot, workingDir)
		if err != nil {
			return nil, fmt.Errorf("working directory must be relative to %s: %w", t.RepoRoot, err)
		}
		dt.WorkingDirectory = targetDir
	} else {
		dt.Directory = t.GetDirectory()
		if t.UsingDocker() {
			workingDir, err := filepath.Rel(dt.Directory, t.workingDir)
			if err != nil {
				return nil, fmt.Errorf("working directory must be relative to %s: %w", dt.Directory, err)
			}
			dt.WorkingDirectory = workingDir
		} else {
			dt.WorkingDirectory = t.workingDir
		}
	}
	dt.AppendArgs("-d", ".")
	if t.Framework != "" {
		dt.AppendArgs("--framework", t.Framework)
	}
	if t.Framework == "terraform" && t.EnableModuleDownload {
		dt.AppendArgs("--download-external-modules", "true")
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
	result.ModuleName = "checkov"
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

func (t *Tool) processResults(result *tools.Result, data *jnode.Node) *tools.Result {
	result.Data = data
	var haveResults bool
	if data.IsArray() {
		// checkov returns an array if it runs more than one check type at a go
		for _, n := range data.Elements() {
			h := t.processCheckResults(result, n)
			haveResults = haveResults || h
		}
	} else {
		haveResults = t.processCheckResults(result, data)
	}
	var summary *jnode.Node
	if haveResults {
		summary = data.Path("summary")
	} else {
		// if checkov has no results then it just returns a summary with
		// no wrapping
		summary = data
	}
	if summary.IsObject() {
		result.AddValue("CHECKOV_VERSION", summary.Path("checkov_version").AsText())
		if rc := summary.Path("resource_count"); !rc.IsMissing() {
			result.AddValue("RESOURCE_COUNT", rc.AsText())
		}
	}
	return result
}

func (t *Tool) processCheckResults(result *tools.Result, e *jnode.Node) bool {
	checkType := e.Path("check_type").AsText()
	results := e.Path("results")
	passedChecks := t.processChecks(result, results.Path("passed_checks"), checkType, true)
	failedChecks := t.processChecks(result, results.Path("failed_checks"), checkType, false)
	updateChecks(results, "passed_checks", passedChecks)
	updateChecks(results, "failed_checks", failedChecks)
	return results.IsObject()
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
		if t.pathTranslationFunc != nil {
			filePath = t.pathTranslationFunc(filePath)
		}
		if len(filePath) > 0 && filePath[0] == '/' {
			// checkov sticks an extra slash at the beginning
			filePath = filePath[1:]
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
