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
	NoInit           bool
	TerraformVersion string
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "tfsec"
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(cmd)
	cmd.Flags().BoolVar(&t.NoInit, "no-init", false, "Don't try and run terraform init on every detected root module first")
	cmd.Flags().StringVar(&t.TerraformVersion, "terraform-version", "", "Use this version of terraform to run init")
}

func (t *Tool) Run() (*tools.Result, error) {
	if !t.NoInit {
		tfInit, err := t.runTerraformInit()
		if err != nil {
			log.Warnf("{warning:terraform init} failed ")
		} else {
			defer tfInit.restore()
		}
	}
	d, err := t.InstallTool(&download.Spec{
		URL: "github.com/tfsec/tfsec",
	})
	if err != nil {
		return nil, err
	}
	customPoliciesDir, err := t.GetCustomPoliciesDir()
	if err != nil {
		return nil, err
	}
	args := []string{"--no-color", "-f", "json"}
	if customPoliciesDir != "" {
		args = append(args, "--external-checks-dir", customPoliciesDir)
	}
	if fi, err := os.Stat(filepath.Join(t.GetDirectory(), "terraform.tfvars")); err == nil && fi.Mode().IsRegular() {
		args = append(args, "--tfvars-file", "terraform.tfvars")
	}
	args = append(args, ".")
	// #nosec G204
	c := exec.Command(d.GetExePath("tfsec-tfsec"), args...)
	c.Dir = t.GetDirectory()
	c.Stderr = os.Stderr
	log.Infof("Running {primary:%s} {secondary:(in %s)}", strings.Join(c.Args, " "), t.GetDirectory())
	output, err := c.Output()
	if util.ExitCode(err) == 1 {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	output = trimOutput(output)
	n, err := jnode.FromJSON(output)
	if err != nil {
		return nil, err
	}

	result := t.parseResults(n)
	result.AddValue("TFSEC_VERSION", d.Version)
	return result, nil
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

func (t *Tool) parseResults(n *jnode.Node) *tools.Result {
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
				FilePath:    filename,
				Line:        r.Path("location").Path("start_line").AsInt(),
				Description: r.Path("description").AsText(),
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
	return &tools.Result{
		Directory: t.GetDirectory(),
		Data:      n,
		Findings:  findings,
		PrintColumns: []string{
			"tool.rule_id",
			"tool.severity",
			"filePath",
			"line",
			"description",
		},
	}
}
