// Copyright 2020 Soluble Inc
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

package root

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/soluble-ai/soluble-cli/cmd/agent"
	"github.com/soluble-ai/soluble-cli/cmd/auth"
	"github.com/soluble-ai/soluble-cli/cmd/aws"
	"github.com/soluble-ai/soluble-cli/cmd/cluster"
	configcmd "github.com/soluble-ai/soluble-cli/cmd/config"
	modelcmd "github.com/soluble-ai/soluble-cli/cmd/model"
	"github.com/soluble-ai/soluble-cli/cmd/query"
	"github.com/soluble-ai/soluble-cli/cmd/version"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/model"
	v "github.com/soluble-ai/soluble-cli/pkg/version"
	"github.com/spf13/cobra"
)

var (
	profile    string
	setProfile string
)

func Command() *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:           "soluble",
		Long:          fmt.Sprintf(`Soluble CLI version %s built %s`, v.Version, v.BuildTime),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if f, _ := cmd.Flags().GetBool("no-color"); f {
				color.NoColor = true
			}
			if f, _ := cmd.Flags().GetBool("quiet"); f {
				log.Level = log.Error
			}
			if f, _ := cmd.Flags().GetBool("debug"); f {
				log.Level = log.Debug
			}
			log.Debugf("Loaded configuration from {primary:%s}", config.ConfigFile)
			if setProfile != "" {
				config.SelectProfile(setProfile)
				config.Save()
			}
			if profile != "" {
				config.SelectProfile(profile)
			}
			log.Infof("Using profile {info:%s} - {primary:%s}", config.GlobalConfig.CurrentProfile,
				config.Config.APIServer)
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "Use this configuration profile (see 'config list-profiles')")
	rootCmd.PersistentFlags().StringVar(&setProfile, "set-profile", "", "Set the current profile to this (and save it.)")
	rootCmd.PersistentFlags().Bool("debug", false, "Run with debug logging")
	rootCmd.PersistentFlags().Bool("quiet", false, "Run with no logging")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output")

	config.Load()
	addBuiltinCommands(rootCmd)
	loadModels()
	for _, model := range model.Models {
		mergeCommands(rootCmd, model.Command.GetCommand().GetCobraCommand(), model.FileName)
	}

	return rootCmd
}

func addBuiltinCommands(rootCmd *cobra.Command) {
	commands := []*cobra.Command{
		auth.Command(),
		agent.Command(),
		aws.Command(),
		cluster.Command(),
		configcmd.Command(),
		modelcmd.Command(),
		version.Command(),
		query.Command(),
	}
	for _, c := range commands {
		rootCmd.AddCommand(c)
	}
}

func loadModels() {
	if err := model.Load(getEmbeddedModelsSource()); err != nil {
		panic(err)
	}
	for _, location := range config.GetModelLocations() {
		source, err := model.GetGitSource(location)
		if err != nil {
			log.Warnf("Could not get models from {info:%s}: {warning:%s}", location, err.Error())
			continue
		}
		err = model.Load(source)
		if err != nil {
			log.Warnf("Could not load models from {info:%s}: {warning:%s}", location, err)
			continue
		}
	}
}

func mergeCommands(root, cmd *cobra.Command, fileName string) {
	for _, existingCommand := range root.Commands() {
		if existingCommand.Use == cmd.Use {
			subCommands := cmd.Commands()
			if len(subCommands) == 0 {
				root.RemoveCommand(existingCommand)
				break
			}
			for _, subCommand := range subCommands {
				mergeCommands(existingCommand, subCommand, fileName)
			}
			return
		}
	}
	root.AddCommand(cmd)
}
