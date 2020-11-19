package root

import (
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/options"
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
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}

{{- if (and (or .Runnable .IsAdditionalHelpTopicCommand) .HasAvailableLocalFlags) }}
  
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
{{- if (and .HasAvailableInheritedFlags .Annotations.ShowInheritedFlags) }}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}
  
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

{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
`)
	rootCmd.AddCommand(options.GetClientOptionsGroupHelpCommand(),
		options.GetPrintOptionsGroupHelpCommand(),
		globalOptionsHelp)
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

func init() {
	cobra.AddTemplateFunc("helpPath", func(cmd *cobra.Command) string {
		path := strings.Split(cmd.CommandPath(), " ")
		helpPath := append([]string{path[0]}, "help")
		if len(path) > 1 {
			helpPath = append(helpPath, path[1:]...)
		}
		return strings.Join(helpPath, " ")
	})
}
