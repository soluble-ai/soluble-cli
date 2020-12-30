package downloadcmd

import (
	"fmt"

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
	opts := options.PrintClientOpts{
		ClientOpts: options.ClientOpts{
			AuthNotRequired: true,
		},
	}
	c := &cobra.Command{
		Use:     "install",
		Short:   "Install a downloadable component",
		Aliases: []string{"reinstall"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			m := download.NewManager()
			var d *download.Download
			var err error
			spec.APIServer = opts.GetAPIClient()
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

func removeCommand() *cobra.Command {
	var (
		name    string
		version string
		all     bool
	)
	c := &cobra.Command{
		Use:   "remove",
		Short: "Remove an installed component",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if all {
				version = ""
			} else if version == "" {
				return fmt.Errorf("either --version or --all must be given")
			}
			m := download.NewManager()
			return m.Remove(name, version)
		},
	}
	flags := c.Flags()
	flags.StringVar(&name, "name", "", "The name of the component to remove")
	flags.StringVar(&version, "version", "", "The version to remove")
	flags.BoolVar(&all, "all", false, "Remove all versions")
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
				v = meta.FindVersion(version, true)
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

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "download",
		Short: "Manage downloaded components",
	}
	c.AddCommand(listCommand())
	c.AddCommand(installCommand())
	c.AddCommand(removeCommand())
	c.AddCommand(getCommand())
	c.AddCommand(printDirCommand())
	return c
}
