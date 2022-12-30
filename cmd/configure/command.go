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
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:     "configure",
		Short:   "Manage IAC component configuration",
		Aliases: []string{"config"},
	}
	c.AddCommand(
		showConfigCmd(),
		listProfilesCmd(),
		updateProfileCmd(),
		setConfigCmd(),
		setProfileCmd(),
		migrateCmd(),
		reconfigCmd(),
	)
	return c
}

func newProfileOpts() *options.PrintOpts {
	return &options.PrintOpts{
		Path: []string{"profiles"},
		Columns: []string{
			"current", "iacProfile", "laceworkProfile", "account", "iacOrganization", "legacyAuth", "iacApiServer",
		},
		Formatters: map[string]print.Formatter{
			"current": func(n *jnode.Node) string {
				if n.AsBool() {
					return "    -->"
				}
				return ""
			},
			"legacyAuth": func(n *jnode.Node) string {
				if n.AsBool() {
					return "*"
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
			Put("laceworkProfile", c.LaceworkProfileName).
			Put("legacyAuth", c.APIToken != "")
		if lp := c.GetLaceworkProfile(); lp != nil {
			m.Put("account", lp.Account)
		} else {
			m.Put("account", c.ConfiguredAccount)
		}
		a.Append(m)
	}
	return n
}

func setProfileCmd() *cobra.Command {
	opts := newProfileOpts()
	var (
		copyFrom  string
		apiServer string
	)
	c := &cobra.Command{
		Use:     "switch-profile <profile>",
		Aliases: []string{"new", "set-profile"},
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
			cfg := config.Get()
			if apiServer != "" {
				cfg.APIServer = apiServer
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
	c.Flags().StringVar(&apiServer, "api-server", "", "Configure APIServer to `url`")
	return c
}

func updateProfileCmd() *cobra.Command {
	opts := newProfileOpts()
	var del bool
	var rename string
	c := &cobra.Command{
		Use:     "update-profile [ <name> ]",
		Short:   "Rename or delete a profile",
		Aliases: []string{"delete", "delete-profile"},
		Args:    cobra.MaximumNArgs(1),
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
			del = del || strings.HasPrefix(cmd.CalledAs(), "delete")
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
		Short:   "Lists IAC profiles",
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
}

func (opts *PrintConfigOpts) PrintConfig() {
	opts.PrintResult(config.Get().PrintableJSON())
}

func showConfigCmd() *cobra.Command {
	var opts PrintConfigOpts
	c := &cobra.Command{
		Use:   "show",
		Short: "Show the configuration of the IAC",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			opts.PrintConfig()
		},
	}
	opts.Register(c)
	return c
}

func setConfigCmd() *cobra.Command {
	var opts PrintConfigOpts
	c := &cobra.Command{
		Use:    "set name value",
		Short:  "Set a CLI configuration parameter",
		Hidden: true,
		Args:   cobra.ExactArgs(2),
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

func migrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Move the legacy config file to its current location if necessary",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.Migrate()
		},
	}
}

func reconfigCmd() *cobra.Command {
	opts := &options.ClientOpts{}
	print := &PrintConfigOpts{}
	c := &cobra.Command{
		Use:     "reconfigure",
		Aliases: []string{"reconfig"},
		Short:   "Modify the current profile",
		Long: `Modify the current profile.
		
Use this command to update the current profile's IAC organization
with the --iac-organization flag.  Use a valid of "" to update the
configuratio to the default IAC organization.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()
			flags := cmd.Flags()
			api, err := opts.GetAPIClient()
			if err != nil {
				return err
			}
			if flags.Changed("iac-organization") {
				api.Organization = opts.APIConfig.Organization
				if api.Organization == "" {
					if err := opts.ConfigureDefaultOrganization(); err != nil {
						return err
					}
				} else {
					if err := opts.ValidateOrganization(); err != nil {
						return err
					}
				}
			}
			cfg.APIServer = api.APIServer
			cfg.APIToken = api.LegacyAPIToken
			cfg.Organization = api.Organization
			if err := config.Save(); err != nil {
				return err
			}
			print.PrintConfig()
			return nil
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.Lookup("api-server").Hidden = false
	flags.Lookup("iac-organization").Hidden = false
	flags.Lookup("disable-tls-verify").Hidden = false
	print.Register(c)
	return c
}
