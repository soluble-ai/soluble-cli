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

package config

import (
	"fmt"
	"sort"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "config",
		Short: "Configure the CLI",
	}
	c.AddCommand(
		showConfigCmd(),
		setConfigCmd(),
		setProfileCmd(),
		listProfilesCmd(),
		updateProfileCmd(),
		migrateCmd())
	return c
}

func newProfileOpts() *options.PrintOpts {
	return &options.PrintOpts{
		Path:    []string{"profiles"},
		Columns: []string{"name", "default", "email", "apiServer", "organization"},
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
		m := jnode.NewObjectNode().Put("name", name).
			Put("default", name == config.GlobalConfig.CurrentProfile).
			Put("email", c.Email).
			Put("apiServer", c.APIServer).
			Put("organization", c.Organization)
		a.Append(m)
	}
	return n
}

func setProfileCmd() *cobra.Command {
	opts := newProfileOpts()
	var (
		name     string
		copyFrom string
	)
	c := &cobra.Command{
		Use:     "set-profile",
		Aliases: []string{"new-profile"},
		Short:   "Set the current profile (or create a new one)",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			config.SelectProfile(name)
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
	c.Flags().StringVar(&name, "name", "", "The name of the profile")
	c.Flags().StringVar(&copyFrom, "copy-from", "", "Copy the profile from another")
	_ = c.MarkFlagRequired("name")
	return c
}

func updateProfileCmd() *cobra.Command {
	opts := newProfileOpts()
	var del bool
	var name string
	var rename string
	c := &cobra.Command{
		Use:   "update-profile",
		Short: "Rename or delete a profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				name = config.GlobalConfig.CurrentProfile
				if name == "" {
					return fmt.Errorf("--name must be given if no current profile is set")
				}
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
	c.Flags().StringVar(&name, "name", "", "The profile to rename or delete")
	c.Flags().StringVar(&rename, "rename", "", "The new name of the profile")
	return c
}

func listProfilesCmd() *cobra.Command {
	opts := newProfileOpts()
	c := &cobra.Command{
		Use:   "list-profiles",
		Short: "Lists the CLI profiles",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			opts.PrintResult(getProfiles(nil))
		},
	}
	opts.Register(c)
	return c
}

type PrintConfigOpts options.PrintOpts

func (opts *PrintConfigOpts) Register(c *cobra.Command) {
	(*options.PrintOpts)(opts).Register(c)
	opts.NoHeaders = true
	opts.Path = []string{"data"}
	opts.Columns = []string{"text"}
}

func (opts *PrintConfigOpts) PrintConfig() {
	n := jnode.NewObjectNode()
	n.PutArray("data").AppendObject().Put("text", config.Config.String())
	(*options.PrintOpts)(opts).PrintResult(n)
}

func showConfigCmd() *cobra.Command {
	var opts PrintConfigOpts
	c := &cobra.Command{
		Use:   "show",
		Short: "Show the configuration of the CLI",
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
