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

package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/blurb"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/soluble-ai/soluble-cli/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ToolOpts struct {
	RunOpts
	Tool                  Interface
	UploadEnabled         bool
	UpdatePR              bool
	PrintAsessment        bool
	PrintResultOpt        bool
	SaveResult            string
	PrintResultValues     bool
	SaveResultValues      string
	DisableCustomPolicies bool
	RepoRoot              string

	customPoliciesDir *string
	config            *Config
}

var _ options.Interface = &ToolOpts{}

func (o *ToolOpts) GetToolOptions() *ToolOpts {
	return o
}

func (o *ToolOpts) GetDirectoryBasedToolOptions() *DirectoryBasedToolOpts {
	return nil
}

func (o *ToolOpts) GetConfig() *Config {
	return o.getConfig(o.RepoRoot)
}

func (o *ToolOpts) getConfig(repoRoot string) *Config {
	if o.config == nil {
		if util.FileExists(filepath.Join(repoRoot, ".soluble", "config.yml")) &&
			!util.FileExists(filepath.Join(repoRoot, ".lacework", "config.yml")) {
			log.Warnf("{info:.soluble/config.yml} is {warning:deprecated}.  Use {info:.lacework/config.yml} instead.")
			o.config = ReadConfig(filepath.Join(repoRoot, ".soluble"))
		} else {
			o.config = ReadConfig(filepath.Join(repoRoot, ".lacework"))
		}
	}
	return o.config
}

func (o *ToolOpts) GetToolHiddenOptions() *options.HiddenOptionsGroup {
	return &options.HiddenOptionsGroup{
		Name: "tool-options",
		Long: "Options for running tools",
		CreateFlagsFunc: func(flags *pflag.FlagSet) {
			flags.BoolVar(&o.DisableCustomPolicies, "disable-custom-policies", false, "Don't use custom policies")
			flags.BoolVar(&o.PrintAsessment, "print-assessment", false, "Print the full assessment response on stderr after an upload")
			flags.BoolVar(&o.PrintResultOpt, "print-result", false, "Print the JSON result from the tool on stderr")
			flags.StringVar(&o.SaveResult, "save-result", "", "Save the JSON reesult from the tool to `file`")
			flags.BoolVar(&o.PrintResultValues, "print-result-values", false, "Print the result values from the tool on stderr")
			flags.StringVar(&o.SaveResultValues, "save-result-values", "", "Save the result values from the tool to `file`")
			flags.BoolVar(&o.UpdatePR, "update-pr", false, "Update this build's pull-request with the resulting assessment")
		},
	}
}

func (o *ToolOpts) Register(c *cobra.Command) {
	// set this now so help shows up, it will be corrected before we print anything
	o.Path = []string{}
	o.AuthNotRequired = true
	o.RunOpts.Register(c)
	flags := c.Flags()
	flags.BoolVar(&o.UploadEnabled, "upload", false, "Upload report to Soluble")
	o.GetToolHiddenOptions().Register(c)
}

func (o *ToolOpts) Validate() error {
	if o.UploadEnabled && o.GetAPIClientConfig().APIToken == "" {
		blurb.SignupBlurb(o, "{info:--upload} requires signing up with {primary:Soluble}.", "")
		return fmt.Errorf("not authenticated with Soluble")
	}
	if o.UpdatePR && !o.UploadEnabled {
		return fmt.Errorf("--update-pr must be used with --upload")
	}
	if o.RepoRoot == "" {
		r, err := inventory.FindRepoRoot(".")
		if err != nil {
			return err
		}
		o.RepoRoot = r
	}
	return nil
}

func (o *ToolOpts) InstallAPIServerArtifact(name, urlPath string) (*download.Download, error) {
	apiClient := o.GetAPIClient()
	m := download.NewManager()
	return m.Install(&download.Spec{
		Name:                       name,
		APIServerArtifact:          urlPath,
		APIServer:                  apiClient,
		LatestReleaseCacheDuration: 1 * time.Minute,
	})
}

func (o *ToolOpts) PrintToolResult(result *Result) {
	o.Columns = result.PrintColumns
	if result.Findings != nil {
		o.Path = []string{}
		o.WideColumns = append(o.WideColumns, "repoPath", "partialFingerprint")
		o.SetFormatter("pass", PassFormatter)
		d, _ := json.Marshal(result.Findings)
		n, _ := jnode.FromJSON(d)
		o.PrintResult(n)
		return
	}
	o.Path = result.PrintPath
	o.Columns = result.PrintColumns
	o.PrintResult(result.Data)
}

func (o *ToolOpts) RunTool(printResult bool) (*Result, error) {
	if err := o.Tool.Validate(); err != nil {
		return nil, err
	}
	result, err := o.Tool.Run()
	if err != nil || result == nil {
		return nil, err
	}
	result.AddValue("TOOL_NAME", o.Tool.Name()).
		AddValue("CLI_VERSION", version.Version).
		AddValue("SOLUBLE_COMMAND_LINE", strings.Join(os.Args, " "))
	if diropts := o.Tool.GetDirectoryBasedToolOptions(); diropts != nil {
		result.Findings.ComputePartialFingerprints(diropts.GetDirectory())
		if o.RepoRoot != "" {
			reldir, err := filepath.Rel(o.RepoRoot, diropts.GetDirectory())
			if err == nil {
				result.AddValue("ASSESSMENT_DIRECTORY", reldir)
			}
		}
	}
	if printResult {
		o.PrintToolResult(result)
	}
	if o.PrintResultOpt {
		p := &print.JSONPrinter{}
		p.PrintResult(os.Stderr, result.Data)
	}
	if o.SaveResult != "" {
		f, err := os.Create(o.SaveResult)
		if err != nil {
			return nil, err
		}
		p := &print.JSONPrinter{}
		p.PrintResult(f, result.Data)
		_ = f.Close()
	}
	if o.PrintResultValues {
		writeResultValues(os.Stderr, result)
	}
	if o.SaveResultValues != "" {
		f, err := os.Create(o.SaveResultValues)
		if err != nil {
			return nil, err
		}
		writeResultValues(f, result)
		_ = f.Close()
	}
	err = result.Report(o.Tool, o.UploadEnabled)
	if err == nil && o.UpdatePR {
		if result.Assessment != nil {
			err = result.Assessment.UpdatePR(o.GetAPIClient())
		}
	}
	return result, err
}

func writeResultValues(w io.Writer, result *Result) {
	for k, v := range result.Values {
		fmt.Fprintf(w, "%s=%s\n", k, v)
	}
}

func (o *ToolOpts) GetCustomPoliciesDir() (string, error) {
	if o.customPoliciesDir != nil {
		return *o.customPoliciesDir, nil
	}
	if o.GetAPIClientConfig().APIToken == "" {
		return "", nil
	}
	d, err := o.InstallAPIServerArtifact(fmt.Sprintf("%s-policies", o.Tool.Name()),
		fmt.Sprintf("/api/v1/org/{org}/rules/%s/rules.tgz", o.Tool.Name()))
	if err != nil {
		return "", err
	}
	// if the directory is empty, then treat that the same as no custom policies
	fs, err := ioutil.ReadDir(d.Dir)
	if err != nil {
		return "", err
	}
	if len(fs) == 0 {
		var zero string
		o.customPoliciesDir = &zero
		log.Infof("{primary:%s} has no custom policies", o.Tool.Name())
	} else {
		o.customPoliciesDir = &d.Dir
	}
	return *o.customPoliciesDir, nil
}
