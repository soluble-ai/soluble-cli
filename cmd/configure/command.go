// Copyright 2022 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package configure

import (
	"fmt"
	"sort"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/api/credentials"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := configureCmd()
	c.AddCommand(
		showConfigCmd(),
		listProfilesCmd(),
		updateProfileCmd(),
		setConfigCmd(),
		setProfileCmd(),
	)
	return c
}

func configureCmd() *cobra.Command {
	configure := &configureCommand{}
	c := &cobra.Command{
		Use:   "configure",
		Short: "Configure the IAC component for use with the lacework CLI",
		Long: `Configure the IAC component for use with the lacework CLI
		
The lacework CLI should already be initialized with "lacework configure".

This command will query the IAC API to determine the user's IAC organization.  If
the organization is already known, it can be specified with the --organization
flag or with the environment variable LW_IAC_ORGANIZATION.

Other subcommands are available, use "configure --help" to list them.`,
		Aliases: []string{"config"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := configure.Run()
			return err
		},
	}
	configure.Register(c)
	return c
}

func newProfileOpts() *options.PrintOpts {
	return &options.PrintOpts{
		Path:    []string{"profiles"},
		Columns: []string{"current", "iacProfile", "laceworkProfile", "domain", "iacApiServer", "iacOrganization"},
		Formatters: map[string]print.Formatter{
			"current": func(n *jnode.Node) string {
				if n.AsBool() {
					return "    -->"
				}
				return ""
			},
		},
	}
}

func getProfiles(names []string) *jnode.Node {
	n := jnode.NewObjectNode()
	a := n.PutArray("profiles")
	if names == nil {
		for name := range config.GlobalConfig.Profiles {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		c := config.GlobalConfig.Profiles[name]
		m := jnode.NewObjectNode().Put("iacProfile", name).
			Put("current", name == config.GlobalConfig.CurrentProfile).
			Put("iacApiServer", c.APIServer).
			Put("iacOrganization", c.Organization).
			Put("laceworkProfile", c.LaceworkProfileName)
		if lp := c.GetLaceworkProfile(); lp != nil {
			m.Put("domain", credentials.GetDomain(lp.Account))
		}
		a.Append(m)
	}
	return n
}

func setProfileCmd() *cobra.Command {
	opts := newProfileOpts()
	var (
		copyFrom string
	)
	c := &cobra.Command{
		Use:     "set-profile <profile>",
		Aliases: []string{"new", "switch-profile"},
		Short:   "Set the current IAC profile (or create a new one)",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			config.SelectProfile(args[0])
			if copyFrom != "" {
				if err := config.CopyProfile(copyFrom); err != nil {
					return err
				}
			}
			if err := config.Save(); err != nil {
				return err
			}
			opts.PrintResult(getProfiles([]string{name}))
			return nil
		},
	}
	opts.Register(c)
	c.Flags().StringVar(&copyFrom, "copy-from", "", "Copy the profile from another")
	_ = c.MarkFlagRequired("name")
	return c
}

func updateProfileCmd() *cobra.Command {
	opts := newProfileOpts()
	var del bool
	var rename string
	c := &cobra.Command{
		Use:   "update-profile [ <name> ]",
		Short: "Rename or delete a profile",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			if len(args) == 1 {
				name = args[0]
			} else {
				name = config.GlobalConfig.CurrentProfile
			}
			if name == "" {
				return fmt.Errorf("no profile specificed and no current profile is set")
			}
			switch {
			case del:
				if config.DeleteProfile(name) {
					if err := config.Save(); err != nil {
						return err
					}
				}
			case rename != "":
				if err := config.RenameProfile(name, rename); err != nil {
					return err
				}
				if err := config.Save(); err != nil {
					return err
				}
			default:
				return fmt.Errorf("either --delete or --rename must be given")
			}
			opts.PrintResult(getProfiles(nil))
			return nil
		},
	}
	opts.Register(c)
	c.Flags().BoolVar(&del, "delete", false, "Delete the named profile")
	c.Flags().StringVar(&rename, "rename", "", "The new name of the profile")
	return c
}

func listProfilesCmd() *cobra.Command {
	opts := newProfileOpts()
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"list-profiles"},
		Short:   "Lists the CLI profiles",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			opts.PrintResult(getProfiles(nil))
		},
	}
	opts.Register(c)
	return c
}

type PrintConfigOpts struct{ options.PrintOpts }

func (opts *PrintConfigOpts) Register(c *cobra.Command) {
	opts.PrintOpts.Register(c)
	opts.NoHeaders = true
	opts.Path = []string{"data"}
	opts.Columns = []string{"text"}
}

func (opts *PrintConfigOpts) PrintConfig() {
	n := jnode.NewObjectNode()
	n.PutArray("data").AppendObject().Put("text", config.Config.String())
	opts.PrintResult(n)
}

func showConfigCmd() *cobra.Command {
	var opts PrintConfigOpts
	c := &cobra.Command{
		Use:   "show",
		Short: "Show the configuration of the CLI",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			log.Infof("Current profile is {primary:%s}", config.GlobalConfig.CurrentProfile)
			opts.PrintConfig()
		},
	}
	opts.Register(c)
	return c
}

func setConfigCmd() *cobra.Command {
	var opts PrintConfigOpts
	c := &cobra.Command{
		Use:   "set name value",
		Short: "Set a CLI configuration parameter",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := config.Set(args[0], args[1])
			if err != nil {
				return err
			}
			defer opts.PrintConfig()
			return config.Save()
		},
	}
	opts.Register(c)
	return c
}
