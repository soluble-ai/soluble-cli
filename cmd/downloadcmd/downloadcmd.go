package downloadcmd

import (
	"fmt"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/spf13/cobra"
)

const githubCom = "github.com"

func listCommand() *cobra.Command {
	opts := options.PrintOpts{
		Path: []string{"data"},
		Columns: []string{
			"Name", "Version", "Dir", "LatestCheckTs",
		},
		WideColumns: []string{
			"URL", "InstallTime",
		},
	}
	c := &cobra.Command{
		Use:   "list",
		Short: "List downloaded components",
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
		name      string
		version   string
		url       string
		reinstall bool
	)
	opts := options.PrintOpts{}
	c := &cobra.Command{
		Use:     "install",
		Short:   "Install a downloadable component",
		Aliases: []string{"reinstall"},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			m := download.NewManager()
			var d *download.Download
			owner, repo := githubRepo(url)
			if owner != "" && version == "" {
				version = "latest"
			}
			if version == "" {
				return fmt.Errorf("--version must be given for plain URL downloads")
			}
			if reinstall || cmd.CalledAs() == "reinstall" {
				if err := m.Remove(name, version); err != nil {
					return err
				}
			}
			if owner != "" {
				d, err = m.InstallGithubRelease(owner, repo, version)
			} else {
				if name == "" || version == "" {
					return fmt.Errorf("--name and --version must be given for plain URL downloads")
				}
				d, err = m.Install(name, version, url)
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
	flags.StringVar(&name, "name", "", "The name of the component to install")
	flags.StringVar(&version, "version", "", "The version to install.  Defaults to the latest release if using github.  Otherwise is required.")
	flags.StringVar(&url, "url", "", "The URL to install. If the URL is in the form github.com/owner/repo then use the github api to install a release")
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

func githubRepo(url string) (string, string) {
	if strings.HasPrefix(url, githubCom) {
		parts := strings.Split(url[len(githubCom)+1:], "/")
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	return "", ""
}

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "download",
		Short: "Manage downloaded components",
	}
	c.AddCommand(listCommand())
	c.AddCommand(installCommand())
	c.AddCommand(removeCommand())
	return c
}
