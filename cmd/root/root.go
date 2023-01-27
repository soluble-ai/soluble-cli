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
	"os"
	"os/exec"
	"strings"

	"github.com/soluble-ai/soluble-cli/cmd/auth"
	"github.com/soluble-ai/soluble-cli/cmd/build"
	"github.com/soluble-ai/soluble-cli/cmd/cfnscan"
	"github.com/soluble-ai/soluble-cli/cmd/cloudscan"
	"github.com/soluble-ai/soluble-cli/cmd/codescan"
	"github.com/soluble-ai/soluble-cli/cmd/configure"
	"github.com/soluble-ai/soluble-cli/cmd/depscan"
	"github.com/soluble-ai/soluble-cli/cmd/downloadcmd"
	"github.com/soluble-ai/soluble-cli/cmd/fingerprint"
	"github.com/soluble-ai/soluble-cli/cmd/helmscan"
	"github.com/soluble-ai/soluble-cli/cmd/imagescan"
	"github.com/soluble-ai/soluble-cli/cmd/inventorycmd"
	"github.com/soluble-ai/soluble-cli/cmd/k8sscan"
	"github.com/soluble-ai/soluble-cli/cmd/kustomizescan"
	"github.com/soluble-ai/soluble-cli/cmd/logincmd"
	modelcmd "github.com/soluble-ai/soluble-cli/cmd/model"
	"github.com/soluble-ai/soluble-cli/cmd/policy"
	"github.com/soluble-ai/soluble-cli/cmd/postcmd"
	"github.com/soluble-ai/soluble-cli/cmd/print"
	"github.com/soluble-ai/soluble-cli/cmd/query"
	"github.com/soluble-ai/soluble-cli/cmd/repoinventory"
	"github.com/soluble-ai/soluble-cli/cmd/secretsscan"
	"github.com/soluble-ai/soluble-cli/cmd/tfscan"
	"github.com/soluble-ai/soluble-cli/cmd/version"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/exit"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/model"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/autoscan"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	v "github.com/soluble-ai/soluble-cli/pkg/version"
	"github.com/spf13/cobra"
)

var (
	ExitFunc = os.Exit
)

func Command() *cobra.Command {
	var (
		profile           string
		workingDir        string
		simulateComponent bool
	)
	// Defer logging until it's been configured
	log.DeferUntilConfigured()

	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:           config.RootCommandName(),
		Long:          fmt.Sprintf("%s version %s", config.RootCommandName(), v.Version),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			log.Configure()
			log.Debugf("Loaded configuration from {primary:%s}", config.ConfigFile)
			if simulateComponent {
				if err := setupSimulatedComponentEnv(); err != nil {
					return err
				}
			}
			if profile != "" {
				if config.SelectProfile(profile) {
					log.Errorf("The IAC profile {info:%s} does not exist, use {primary:%s configure switch-profile %s} to create it",
						profile, config.CommandInvocation(), profile)
					return fmt.Errorf("the IAC profile %s does not exist", profile)
				}
			}
			if workingDir != "" {
				if err := os.Chdir(workingDir); err != nil {
					return err
				}
			}
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if exit.Code != 0 {
				if exit.Func != nil {
					exit.Func()
				}
				ExitFunc(exit.Code)
			}
		},
		Version: v.Version,
	}

	flags := rootCmd.PersistentFlags()
	flags.StringVar(&profile, "iac-profile", "", "Use this configuration profile (see 'config list-profiles')")
	log.AddFlags(flags)
	flags.StringVar(&workingDir, "working-dir", "", "Change the working dir to `dir` before running")
	flags.Lookup("working-dir").Hidden = true
	flags.BoolVar(&simulateComponent, "simulate-component", false, "Simulate running as a component")
	flags.Lookup("simulate-component").Hidden = true

	config.Load()
	addBuiltinCommands(rootCmd)
	loadModels()
	for _, model := range model.Models {
		mergeCommands(rootCmd, model.Command.GetCommand().GetCobraCommand(), model)
	}
	setupHelp(rootCmd)
	return rootCmd
}

func SetAnnotation(cmd *cobra.Command, name, value string) {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	cmd.Annotations[name] = value
}

func addBuiltinCommands(rootCmd *cobra.Command) {
	// make this hidden for now
	checkovCommand := tools.CreateCommand(&checkov.Tool{})
	checkovCommand.Hidden = true
	rootCmd.AddCommand(
		auth.Command(),
		modelcmd.Command(),
		version.Command(),
		query.Command(),
		downloadcmd.Command(),
		postcmd.Command(),
		imagescan.Command(),
		inventorycmd.Command(),
		build.Command(),
		depscan.Command(),
		k8sscan.Command(),
		helmscan.Command(),
		tfscan.Command(),
		secretsscan.Command(),
		cfnscan.Command(),
		tools.CreateCommand(&autoscan.Tool{}),
		checkovCommand,
		codescan.Command(),
		cloudscan.Command(),
		fingerprint.Command(),
		earlyAccessCommand(),
		repoinventory.Command(),
		kustomizescan.Command(),
		print.Command(),
		configure.Command(),
		policy.Command(),
	)
	if !config.IsRunningAsComponent() {
		rootCmd.AddCommand(logincmd.Command())
	}
}

func loadModels() {
	if err := model.Load(getEmbeddedModelsSource()); err != nil {
		panic(err)
	}
	if os.Getenv("SOLUBLE_DISABLE_CLI_MODELS") != "" {
		return
	}
	sources := make(chan model.Source)
	for _, loc := range config.GetModelLocations() {
		location := loc
		go func() {
			source, err := model.GetGitSource(location)
			if err != nil {
				log.Warnf("Could not get models from {info:%s}: {warning:%s}", location, err.Error())
			} else {
				sources <- source
			}
		}()
	}
	for i := len(config.GetModelLocations()); i > 0; i-- {
		source := <-sources
		if source != nil {
			err := model.Load(source)
			if err != nil {
				log.Warnf("Could not load models from {info:%s}: {warning:%s}", source.String(), err)
				continue
			}
		}
	}
}

func mergeCommands(root, cmd *cobra.Command, m *model.Model) {
	for _, existingCommand := range root.Commands() {
		if existingCommand.Use == cmd.Use {
			subCommands := cmd.Commands()
			if len(subCommands) == 0 {
				root.RemoveCommand(existingCommand)
				break
			}
			if existingCommand.Short == "" && cmd.Short != "" {
				// take short from model if present
				existingCommand.Short = cmd.Short
			}
			for _, subCommand := range subCommands {
				mergeCommands(existingCommand, subCommand, m)
			}
			return
		}
	}
	if !m.Source.IsEmbedded() {
		cmd.Short += " (" + m.Source.String() + ")"
	}
	root.AddCommand(cmd)
}

func setupSimulatedComponentEnv() error {
	if config.IsRunningAsComponent() {
		return nil
	}
	var (
		account   string
		apiKey    string
		apiSecret string
		apiToken  string
		err       error
	)
	account, err = runLacework("configure", "show", "account")
	if err == nil {
		apiKey, err = runLacework("configure", "show", "api_key")
	}
	if err == nil {
		apiSecret, err = runLacework("configure", "show", "api_secret")
	}
	if err == nil {
		apiToken, err = runLacework("access-token")
	}
	if err != nil {
		return err
	}
	log.Infof("Setting up environment to simulate component execution with account {info:%s}", account)
	os.Setenv("LW_ACCOUNT", account)
	os.Setenv("LW_API_KEY", apiKey)
	os.Setenv("LW_API_SECRET", apiSecret)
	os.Setenv("LW_API_TOKEN", apiToken)
	os.Setenv("LW_COMPONENT_NAME", "iac")
	return nil
}

func runLacework(args ...string) (string, error) {
	c := exec.Command("lacework", args...)
	out, err := c.Output()
	if err != nil {
		return "", err
	}
	s := string(out)
	nl := strings.IndexRune(s, '\n')
	if nl > 0 {
		s = s[0:nl]
	}
	return strings.TrimSpace(s), err
}
