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

package tfsec

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
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
	NoInit           bool
	TerraformVersion string
	TerraformCommand string

	extraArgs tools.ExtraArgs
}

var v0_39_38 = version.Must(version.NewVersion("0.39.38"))

var _ tools.Single = &Tool{}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "tfsec",
		Short: "Scan terraform for security vulnerabilities",
		Args:  t.extraArgs.ArgsValue(),
	}
}

func (t *Tool) Name() string {
	return "tfsec"
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(cmd)
	cmd.Flags().BoolVar(&t.NoInit, "no-init", false, "Don't try and run terraform init on every detected root module first")
	cmd.Flags().StringVar(&t.TerraformVersion, "terraform-version", "", "Use this version of terraform to run init")
	cmd.Flags().StringVar(&t.TerraformCommand, "terraform-command", "", "Use `command` for terraform instead of downloading a version.")
}

func (t *Tool) Run() (*tools.Result, error) {
	result := &tools.Result{
		Directory:   t.GetDirectory(),
		IACPlatform: "terraform",
	}
	if !t.NoInit {
		tfInit, err := t.runTerraformInit()
		if err != nil {
			log.Warnf("{warning:terraform init} failed ")
			result.AddValue("TERRAFORM_INIT_FAILED", "true")
		} else {
			defer tfInit.restore()
		}
	}
	d, err := t.InstallTool(&download.Spec{
		URL: "github.com/aquasecurity/tfsec",
	})
	if err != nil {
		return nil, err
	}
	tfsecVersion := d.Version
	if tfsecVersion == "" {
		tfsecVersion = getTfsecVersion(d.GetExePath("aquasecurity-tfsec"))
	}
	result.AddValue("TFSEC_VERSION", tfsecVersion)
	customPoliciesDir, err := t.GetCustomPoliciesDir()
	if err != nil {
		return nil, err
	}
	args := []string{"--no-color", "-f", "json"}
	if tfsecVersion != "" {
		v, err := version.NewSemver(tfsecVersion)
		if err == nil && v.GreaterThanOrEqual(v0_39_38) {
			args = append(args, "--include-ignored")
			args = append(args, "--include-passed")
		}
	}
	if customPoliciesDir != "" {
		args = append(args, "--custom-check-dir", customPoliciesDir)
	}
	args = t.addTfVarsFileArg(args, "terraform.tfvars")
	args = t.addTfVarsFileArg(args, "terraform.tfvars.json")
	args = t.addAutoTfVarsFiles(args)
	args = append(args, t.extraArgs...)
	args = append(args, ".")
	// #nosec G204
	c := exec.Command(d.GetExePath("aquasecurity-tfsec"), args...)
	c.Dir = t.GetDirectory()
	c.Stderr = os.Stderr
	exec := t.ExecuteCommand(c)
	result.ExecuteResult = exec
	if !exec.ExpectExitCode(0, 1) {
		return result, nil
	}
	exec.Output = trimOutput(exec.Output)
	n, ok := exec.ParseJSON()
	if !ok {
		return result, nil
	}
	t.parseResults(result, n)
	return result, nil
}

func (t *Tool) addTfVarsFileArg(args []string, name string) []string {
	if fi, err := os.Stat(filepath.Join(t.GetDirectory(), name)); err == nil && fi.Mode().IsRegular() {
		args = append(args, "--tfvars-file", name)
	}
	return args
}

func (t *Tool) addAutoTfVarsFiles(args []string) []string {
	err := filepath.WalkDir(t.GetDirectory(), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && t.GetDirectory() != path {
			return filepath.SkipDir
		}
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".auto.tfvars") || strings.HasSuffix(path, ".auto.tfvars.json") {
			args = append(args, "--tfvars-file", d.Name())
		}
		return nil
	})
	if err != nil {
		log.Warnf("Could not read directory {warning:%s}", err)
	}
	return args
}

func trimOutput(output []byte) []byte {
	// tfsec unhelpfully logs stuff before json, so skip over that
	for i := range output {
		if output[i] == '{' {
			if i > 0 {
				log.Warnf("tfsec warning:\n{warning:%s}", strings.TrimSpace(string(output[0:i])))
			}
			return output[i:]
		}
	}
	_, _ = os.Stderr.Write(output)
	log.Warnf("{warning:tfsec} did not output JSON")
	return output
}

func (t *Tool) parseResults(result *tools.Result, n *jnode.Node) {
	dir := t.GetDirectory()
	results := n.Path("results")
	var findings []*assessments.Finding
	if results.Size() > 0 {
		for _, r := range n.Path("results").Elements() {
			loc := r.Path("location")
			filename := loc.Path("filename").AsText()
			if filename != "" && filepath.IsAbs(filename) {
				f, err := filepath.Rel(dir, filename)
				if err == nil {
					loc.Put("filename", f)
					filename = f
				}
			}
			if t.IsExcluded(filename) {
				continue
			}
			findings = append(findings, &assessments.Finding{
				FilePath:      filename,
				Line:          r.Path("location").Path("start_line").AsInt(),
				Description:   r.Path("description").AsText(),
				GeneratedFile: strings.HasPrefix(filepath.ToSlash(filename), ".terraform/modules/"),
				Pass:          r.Path("status").AsInt() == 1,
				Tool: map[string]string{
					"severity": r.Path("severity").AsText(),
					"rule_id":  r.Path("rule_id").AsText(),
				},
			})
		}
		results = util.RemoveJNodeElementsIf(results, func(e *jnode.Node) bool {
			return t.IsExcluded(e.Path("location").Path("filename").AsText())
		})
		n.Put("results", results)
	}
	result.Data = n
	result.Findings = findings
}

func getTfsecVersion(tfsecPath string) string {
	c := exec.Command(tfsecPath, "-v")
	out, err := c.Output()
	if err == nil {
		var v *version.Version
		v, err = version.NewVersion(strings.TrimSpace(string(out)))
		if err == nil {
			log.Infof("{primary:tfsec} is version {info:%s}", v.Original())
			return v.Original()
		}
	}
	log.Warnf("Could not determine {primary:tfsec} version - {warning:%s}", err)
	return ""
}
