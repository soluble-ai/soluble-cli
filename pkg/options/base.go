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

package options

import (
	"fmt"

	"github.com/spf13/cobra"
)

type Interface interface {
	Register(cmd *cobra.Command)
	SetContextValues(context map[string]string)
}

type HiddenOptionsGroup struct {
	Use         string
	Description string
	OptionNames []string
}

func AddHiddenOptionsGroup(cmd *cobra.Command, group *HiddenOptionsGroup) {
	for _, name := range group.OptionNames {
		_ = cmd.Flags().MarkHidden(name)
	}
	help := &cobra.Command{
		Use:   group.Use,
		Short: fmt.Sprintf("Show help for flags that %s", group.Description),
		Long:  fmt.Sprintf("These flags %s", group.Description),
	}
	for _, name := range group.OptionNames {
		flag := cmd.Flags().Lookup(name)
		if flag != nil {
			copy := *flag
			copy.Hidden = false
			help.Flags().AddFlag(&copy)
		}
	}
	// suppress the default help flag
	help.Flags().BoolP("help", "h", false, "Display help")
	_ = help.Flags().MarkHidden("help")
	help.SetHelpTemplate(`{{.Long}}

{{.LocalFlags.FlagUsages}}
`)
	cmd.AddCommand(help)
}
