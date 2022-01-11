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
	PrintResultOpt        bool
	SaveResult            string
	PrintResultValues     bool
	SaveResultValues      string
	DisableCustomPolicies bool
	RepoRoot              string
	PrintFingerprints     bool
	SaveFingerprints      string

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
		oldConfig := filepath.Join(repoRoot, ".soluble", "config.yml")
		newConfig := filepath.Join(repoRoot, ".lacework", "config.yml")
		if util.FileExists(oldConfig) && !util.FileExists(newConfig) {
			log.Warnf("{info:%s} is {warning:deprecated}.  Use {info:%s} instead.",
				oldConfig, newConfig)
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
			flags.BoolVar(&o.PrintResultOpt, "print-result", false, "Print the JSON result from the tool on stderr")
			flags.StringVar(&o.SaveResult, "save-result", "", "Save the JSON reesult from the tool to `file`")
			flags.BoolVar(&o.PrintResultValues, "print-result-values", false, "Print the result values from the tool on stderr")
			flags.StringVar(&o.SaveResultValues, "save-result-values", "", "Save the result values from the tool to `file`")
			flags.BoolVar(&o.PrintFingerprints, "print-fingerprints", false, "Print fingerprints on stderr before uploading results")
			flags.StringVar(&o.SaveFingerprints, "save-fingerprints", "", "Save finding fingerprints to `file`")
		},
	}
}

func (o *ToolOpts) Register(c *cobra.Command) {
	o.Path = []string{}
	o.Columns = []string{
		"sid", "severity", "pass", "title", "filePath", "line",
	}
	o.SetFormatter("pass", PassFormatter)
	// if not uploaded these columns will be empty, so make that a little easier to see
	o.SetFormatter("sid", MissingFormatter)
	o.SetFormatter("severity", MissingFormatter)
	o.AuthNotRequired = true
	o.RunOpts.Register(c)
	flags := c.Flags()
	flags.BoolVar(&o.UploadEnabled, "upload", true, "Upload report to Soluble.  Use --upload=false to disable.")
	o.GetToolHiddenOptions().Register(c)
}

func (o *ToolOpts) Validate() error {
	if o.UploadEnabled && o.GetAPIClientConfig().APIToken == "" {
		blurb.SignupBlurb(o, "{info:--upload} requires signing up with {primary:Soluble}.", "")
		return fmt.Errorf("not authenticated with Soluble")
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

func (o *ToolOpts) RunTool() (Results, error) {
	if err := o.Tool.Validate(); err != nil {
		return nil, err
	}
	var (
		results Results
		err     error
	)
	if s, ok := o.Tool.(Single); ok {
		var r *Result
		r, err = s.Run()
		if r != nil {
			results = Results{r}
		}
	} else if c, ok := o.Tool.(Consolidated); ok {
		results, err = c.RunAll()
	}
	for _, result := range results {
		rerr := o.processResult(result)
		if rerr != nil {
			// processResult only fails if the upload failed, and if that
			// fais then it's likely that nothing is going to work
			return nil, rerr
		}
	}
	return results, err
}

func (o *ToolOpts) processResult(result *Result) error {
	result.AddValue("TOOL_NAME", o.Tool.Name()).
		AddValue("CLI_VERSION", version.Version).
		AddValue("SOLUBLE_COMMAND_LINE", strings.Join(os.Args, " "))
	if result.Directory != "" {
		result.UpdateFileFingerprints()
		if o.RepoRoot != "" {
			reldir, err := filepath.Rel(o.RepoRoot, result.Directory)
			if err == nil && !strings.HasPrefix(reldir, "..") {
				result.AddValue("ASSESSMENT_DIRECTORY", reldir)
			}
		}
	}
	if o.PrintFingerprints || o.SaveFingerprints != "" {
		d, err := json.Marshal(result.FileFingerprints)
		util.Must(err)
		n, err := jnode.FromJSON(d)
		util.Must(err)
		if o.PrintFingerprints {
			p := &print.JSONPrinter{}
			p.PrintResult(os.Stderr, n)
		}
		if o.SaveFingerprints != "" {
			p := &print.JSONPrinter{}
			f, err := os.Create(o.SaveFingerprints)
			if err != nil {
				log.Warnf("Could not save fingerprints: {warning:%s}", err)
			} else {
				p.PrintResult(f, n)
				_ = f.Close()
			}
		}
	}
	if o.PrintResultOpt {
		p := &print.JSONPrinter{}
		p.PrintResult(os.Stderr, result.Data)
	}
	if o.SaveResult != "" {
		f, err := os.Create(o.SaveResult)
		if err != nil {
			return err
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
			return err
		}
		writeResultValues(f, result)
		_ = f.Close()
	}
	if o.UploadEnabled {
		if err := result.Upload(o.GetAPIClient(), o.GetOrganization(), o.Tool.Name()); err != nil {
			return err
		}
	}
	return nil
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
