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

package downloadcmd

import (
	"fmt"
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/spf13/cobra"
)

func listCommand() *cobra.Command {
	opts := options.PrintOpts{
		Path: []string{"data"},
		Columns: []string{
			"Name", "Version", "Dir", "LatestCheckTs+",
		},
		WideColumns: []string{
			"URL", "InstallTime",
		},
	}
	c := &cobra.Command{
		Use:   "list",
		Short: "List downloaded components",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			m := download.NewManager()
			result := m.List()
			n := jnode.NewObjectNode()
			a := n.PutArray("data")
			for _, r := range result {
				for _, v := range r.Installed {
					m, err := print.ToResult(v)
					if err != nil {
						return err
					}
					if v.Version == r.LatestVersion {
						m.Put("LatestCheckTs", r.LatestCheckTime.String())
					}
					a.Append(m)
				}
			}
			opts.PrintResult(n)
			return nil
		},
	}
	opts.Register(c)
	return c
}

func installCommand() *cobra.Command {
	var (
		spec      download.Spec
		reinstall bool
	)
	var opts options.PrintClientOpts
	c := &cobra.Command{
		Use:     "install",
		Short:   "Install a downloadable component",
		Aliases: []string{"reinstall"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			m := download.NewManager()
			var d *download.Download
			api, err := opts.GetAPIClient()
			if err != nil {
				return err
			}
			spec.APIServer = api
			if spec.RequestedVersion == "" {
				n, err := getDefaultVersion(&opts, spec.Name)
				if err != nil {
					return err
				}
				spec.RequestedVersion = n.Path("version").AsText()
			}
			if cmd.CalledAs() == "reinstall" || reinstall {
				d, err = m.Reinstall(&spec)
			} else {
				d, err = m.Install(&spec)
			}
			if err != nil {
				return err
			}
			result, err := print.ToResult(d)
			if err != nil {
				return err
			}
			opts.PrintResult(result)
			return nil
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVar(&spec.Name, "name", "", "The name of the component to install")
	flags.StringVar(&spec.RequestedVersion, "version", "", "The version to install.  Defaults to the latest release if using github.  Otherwise is required.")
	flags.StringVar(&spec.URL, "url", "", "The URL to install. If the URL is in the form github.com/owner/repo then use the github api to install a release")
	flags.StringVar(&spec.APIServerArtifact, "soluble-artifact", "", "Install an artifact from Soluble")
	flags.BoolVar(&reinstall, "reinstall", false, "Reinstall the component")
	return c
}

func getCommand() *cobra.Command {
	var name string
	opts := options.PrintOpts{}
	c := &cobra.Command{
		Use:   "get",
		Short: "Get details of a downloaded component",
		RunE: func(cmd *cobra.Command, args []string) error {
			m := download.NewManager()
			meta := m.GetMeta(name)
			if meta == nil {
				log.Errorf("The component {warning:%s} is not installed", name)
				return fmt.Errorf("component not installed")
			}
			n, _ := print.ToResult(meta)
			opts.PrintResult(n)
			return nil
		},
	}
	opts.Register(c)
	c.Flags().StringVar(&name, "name", "", "The name of the component to display")
	_ = c.MarkFlagRequired("name")
	return c
}

func printDirCommand() *cobra.Command {
	var (
		name    string
		version string
	)
	opts := options.PrintOpts{}
	c := &cobra.Command{
		Use:   "print-dir",
		Short: "Print the installation directory a downloaded component",
		RunE: func(cmd *cobra.Command, args []string) error {
			m := download.NewManager()
			meta := m.GetMeta(name)
			if meta == nil {
				return fmt.Errorf("component not installed")
			}
			var v *download.Download
			if version == "" {
				v = meta.FindLatestOrLastInstalledVersion()
			} else {
				v = meta.FindVersion(version, 0, true)
			}
			if v == nil {
				return fmt.Errorf("version not found")
			}
			fmt.Println(v.Dir)
			return nil
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			log.Level = log.Error
		},
	}
	opts.Register(c)
	c.Flags().StringVar(&name, "name", "", "The name of the component to display")
	c.Flags().StringVar(&version, "version", "", "The version.  If unspecified, the latest version.")
	_ = c.MarkFlagRequired("name")
	return c
}

func getDefaultVersion(opts *options.PrintClientOpts, name string) (*jnode.Node, error) {
	return opts.GetUnauthenticatedAPIClient().Get(fmt.Sprintf("cli/tools/%s/config", name))
}

func getDefaultVersionCommand() *cobra.Command {
	opts := &options.PrintClientOpts{}
	var name string
	c := &cobra.Command{
		Use:   "get-default-version",
		Short: "Print the default version of a tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := getDefaultVersion(opts, name)
			if err != nil {
				return err
			}
			opts.PrintResult(n)
			return nil
		},
	}
	opts.Register(c)
	c.Flags().StringVar(&name, "name", "", "The name of the tool")
	_ = c.MarkFlagRequired("name")
	return c
}

func cleanCommand() *cobra.Command {
	var (
		all  bool
		name string
	)
	c := &cobra.Command{
		Use:   "clean",
		Short: "Remove all but the latest version of all downloads",
		RunE: func(cmd *cobra.Command, args []string) error {
			m := download.NewManager()
			for _, dm := range m.List() {
				if name != "" && name != dm.Name {
					continue
				}
				if all {
					log.Infof("Removing {primary:%s}", dm.Dir)
					if err := os.RemoveAll(dm.Dir); err != nil {
						return err
					}
				} else {
					for _, d := range dm.Installed {
						if d.Version != dm.LatestVersion {
							if err := m.Remove(dm.Name, d.Version); err != nil {
								return err
							}
						}
					}
				}
			}
			return nil
		},
	}
	flags := c.Flags()
	flags.BoolVar(&all, "all", false, "Remove all versions including latest")
	flags.StringVar(&name, "name", "", "Remove just a specific download")
	return c
}

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "download",
		Short: "Manage downloaded components",
	}
	c.AddCommand(
		listCommand(),
		installCommand(),
		getCommand(),
		printDirCommand(),
		getDefaultVersionCommand(),
		cleanCommand(),
	)
	return c
}
