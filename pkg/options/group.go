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
		f.Usage = fmt.Sprintf("%s.  See help-%s for examples.", f.Usage, group.Name)
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

{{.LocalFlags.FlagUsagesWrapped 100 | trimTrailingWhitespaces}}{{ if .HasExample }}
{{.Example}}
{{end}}
`)
	return c
}
