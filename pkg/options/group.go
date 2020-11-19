package options

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type HiddenOptionsGroup struct {
	Name            string
	Long            string
	Example         string
	CreateFlagsFunc func(*pflag.FlagSet)
}

func (group *HiddenOptionsGroup) Register(cmd *cobra.Command) {
	flags := &pflag.FlagSet{}
	group.CreateFlagsFunc(flags)
	flags.VisitAll(func(f *pflag.Flag) {
		f.Hidden = true
		cmd.Flags().AddFlag(f)
	})
}

func (group *HiddenOptionsGroup) GetHelpCommand() *cobra.Command {
	c := &cobra.Command{
		Use:     fmt.Sprintf("help-%s", group.Name),
		Long:    group.Long,
		Example: group.Example,
	}
	group.CreateFlagsFunc(c.Flags())
	c.SetHelpTemplate(`{{.Long}}

{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{ if .HasExample }}
{{.Example}}
{{end}}
`)
	return c
}
