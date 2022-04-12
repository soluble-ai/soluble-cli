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
	"os"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/repotree"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/soluble-ai/soluble-cli/pkg/version"
	"github.com/spf13/cobra"
)

type ToolOpts struct {
	RunOpts
	Tool       Interface
	RepoRoot   string
	ConfigFile string

	config      *Config
	repoRootSet bool
}

var _ options.Interface = &ToolOpts{}

func (o *ToolOpts) GetToolOptions() *ToolOpts {
	return o
}

func (o *ToolOpts) GetConfig() *Config {
	return o.getConfig(o.RepoRoot)
}

func (o *ToolOpts) getConfig(repoRoot string) *Config {
	if o.config == nil {
		if o.ConfigFile != "" {
			o.config = ReadConfigFile(o.ConfigFile)
		} else {
			oldConfig := filepath.Join(repoRoot, ".soluble", "config.yml")
			newConfig := filepath.Join(repoRoot, ".lacework", "config.yml")
			if util.FileExists(oldConfig) && !util.FileExists(newConfig) {
				log.Warnf("{info:%s} is {warning:deprecated}.  Use {info:%s} instead.",
					oldConfig, newConfig)
				o.config = ReadConfigFile(oldConfig)
			} else {
				o.config = ReadConfigFile(newConfig)
			}
		}
	}
	return o.config
}

func (o *ToolOpts) Register(cmd *cobra.Command) {
	o.RunOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVar(&o.ConfigFile, "config-file", "", "Read tool configuration from `file`, overriding the default config file search.")
	_ = flags.MarkHidden("config-file")
}

func (o *ToolOpts) Validate() error {
	if o.RepoRoot == "" && !o.repoRootSet {
		r, err := repotree.FindRepoRoot(".")
		if err != nil {
			return err
		}
		o.RepoRoot = r
	}
	return nil
}

func (o *ToolOpts) GetStandardXCPValues() map[string]string {
	return map[string]string{
		"CLI_VERSION":          version.Version,
		"SOLUBLE_COMMAND_LINE": strings.Join(os.Args, " "),
		"TOOL_NAME":            o.Tool.Name(),
	}
}
