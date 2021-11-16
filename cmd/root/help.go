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

package root

import (
	"fmt"
	"strings"

	"github.com/mitchellh/go-wordwrap"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func setupHelp(rootCmd *cobra.Command) {
	rootCmd.SetHelpCommand(helpCommand(rootCmd))
	rootCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
{{.NameAndAliases}}{{end}}{{if .HasExample}}
  
Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}
  
Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short | wrap (plus .NamePadding 3) 100}}{{if gt (len .Aliases) 0}}
    aliases: {{.Aliases | joinCommas}}{{end}}{{end}}{{end}}{{end}}

{{- if (and (or .Runnable .IsAdditionalHelpTopicCommand) .HasAvailableLocalFlags) }}
  
Flags:
{{.LocalFlags.FlagUsagesWrapped 100 | trimTrailingWhitespaces}}{{end}}
{{- if (and .HasAvailableInheritedFlags .Annotations.ShowInheritedFlags) }}

Global Flags:
{{.InheritedFlags.FlagUsagesWrapped 100 | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}
  
Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
  
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
{{- if (and .Runnable (not .Annotations.ShowInheritedFlags)) }}

Some options have been hidden, use "{{ helpPath $ }} -a" to display all options{{end}}

`)
	globalOptionsHelp := &cobra.Command{
		Use:  "help-global-options",
		Long: "Global options",
		Annotations: map[string]string{
			"ShowInheritedFlags": "1",
		},
	}
	globalOptionsHelp.SetHelpTemplate(`{{.Long}}

{{.InheritedFlags.FlagUsagesWrapped 100 | trimTrailingWhitespaces}}
`)
	rootCmd.AddCommand(
		options.GetClientOptionsGroupHelpCommand(),
		options.GetPrintOptionsGroupHelpCommand(),
		globalOptionsHelp,
		(&tools.ToolOpts{}).GetToolHiddenOptions().GetHelpCommand(),
		(&tools.RunOpts{}).GetRunHiddenOptions().GetHelpCommand(),
		(&tools.DirectoryBasedToolOpts{}).GetDirectoryBasedHiddenOptions().GetHelpCommand(),
	)
}

func helpCommand(rootCmd *cobra.Command) *cobra.Command {
	var (
		allOptions bool
	)
	c := &cobra.Command{
		Use:   "help",
		Short: "Help about any command",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _, err := rootCmd.Find(args)
			if err != nil {
				return err
			}
			if target.Annotations == nil {
				target.Annotations = map[string]string{}
			}
			if allOptions {
				target.LocalFlags().VisitAll(func(flag *pflag.Flag) {
					flag.Hidden = false
				})
				target.Annotations["ShowInheritedFlags"] = "1"
			}
			return target.Help()
		},
		Annotations: map[string]string{"ShowInheritedFlags": ""},
	}
	flags := c.Flags()
	flags.BoolVarP(&allOptions, "all-options", "a", false, "Show help for all options")
	return c
}

func wrap(indent, length int, s string) string {
	p := strings.Repeat(" ", indent)
	lines := strings.Split(wordwrap.WrapString(s, uint(length-indent)), "\n")
	for i := range lines {
		if i > 0 {
			lines[i] = fmt.Sprintf("%s%s", p, lines[i])
		}
	}
	return strings.Join(lines, "\n")
}

func init() {
	cobra.AddTemplateFunc("helpPath", func(cmd *cobra.Command) string {
		path := strings.Split(cmd.CommandPath(), " ")
		helpPath := append([]string{path[0]}, "help")
		if len(path) > 1 {
			helpPath = append(helpPath, path[1:]...)
		}
		return strings.Join(helpPath, " ")
	})
	cobra.AddTemplateFunc("joinCommas", func(vals []string) string {
		return strings.Join(vals, ",")
	})
	cobra.AddTemplateFunc("wrap", wrap)
	cobra.AddTemplateFunc("plus", func(i, j int) int { return i + j })
}
