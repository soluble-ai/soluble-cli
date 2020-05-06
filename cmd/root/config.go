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
	"sort"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func ConfigCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "config",
		Short: "Configure the CLI",
	}
	c.AddCommand(showConfigCmd())
	c.AddCommand(setConfigCmd())
	c.AddCommand(setProfileCmd())
	c.AddCommand(listProfilesCmd())
	c.AddCommand(updateProfileCmd())
	return c
}

func newProfileOpts() *options.PrintOpts {
	return &options.PrintOpts{
		Path:    []string{"profiles"},
		Columns: []string{"name", "default", "email", "apiServer", "organization"},
	}
}

func getProfiles() *jnode.Node {
	n := jnode.NewObjectNode()
	a := n.PutArray("profiles")
	names := []string{}
	for name := range config.GlobalConfig.Profiles {
		names = append(names, name)
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
	var name string
	c := &cobra.Command{
		Use:   "set-profile",
		Short: "Set the current profile",
		Run: func(cmd *cobra.Command, args []string) {
			config.SelectProfile(name)
			config.Save()
			opts.PrintResult(getProfiles())
		},
	}
	opts.Register(c)
	c.Flags().StringVar(&name, "name", "", "The name of the profile")
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
					config.Save()
				}
			case rename != "":
				if err := config.RenameProfile(name, rename); err != nil {
					return err
				}
				config.Save()
			default:
				return fmt.Errorf("either --delete or --rename must be given")
			}
			opts.PrintResult(getProfiles())
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
		Run: func(cmd *cobra.Command, args []string) {
			opts.PrintResult(getProfiles())
		},
	}
	opts.Register(c)
	return c
}

func showConfigCmd() *cobra.Command {
	var opts options.PrintOpts
	c := &cobra.Command{
		Use:   "show",
		Short: "Show the configuration of the CLI",
		Run: func(cmd *cobra.Command, args []string) {
			printConfig(&opts)
		},
	}
	opts.Register(c)
	return c
}

func setConfigCmd() *cobra.Command {
	var opts options.PrintOpts
	c := &cobra.Command{
		Use:   "set name value",
		Short: "Set a CLI configuration parameter",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := config.Set(args[0], args[1])
			if err != nil {
				return err
			}
			config.Save()
			printConfig(&opts)
			return nil
		},
	}
	opts.Register(c)
	return c
}

func printConfig(opts *options.PrintOpts) {
	log.Infof("Current profile {primary:%s} loaded from {info:%s}", config.GlobalConfig.CurrentProfile, config.ConfigFile)
	fmt.Fprintln(opts.GetOutput(), config.Config.String())
}
