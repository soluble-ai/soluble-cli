package iacscan

import (
	"fmt"

	"github.com/soluble-ai/soluble-cli/pkg/tools"
	cfnpythonlint "github.com/soluble-ai/soluble-cli/pkg/tools/cfn-python-lint"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/soluble-ai/soluble-cli/pkg/tools/cloudformationguard"
	"github.com/soluble-ai/soluble-cli/pkg/tools/terrascan"
	"github.com/soluble-ai/soluble-cli/pkg/tools/tfsec"
	"github.com/spf13/cobra"
)

func createCommand(tool tools.Interface) *cobra.Command {
	opts := &tools.ToolOpts{}
	c := &cobra.Command{
		Use:   tool.Name(),
		Short: fmt.Sprintf("Scan infrastructure-as-code with %s", tool.Name()),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if u, ok := tool.(tools.RunsInDirectory); ok {
				d, _ := cmd.Flags().GetString("directory")
				if d == "" {
					return fmt.Errorf("%s requires --directory", tool.Name())
				}
				u.SetDirectory(d)
			}
			if u, ok := tool.(tools.RunsWithAPIClient); ok {
				u.SetAPIClient(opts.GetAPIClient())
			}
			return opts.RunTool(tool)
		},
	}
	opts.Register(c)
	flags := c.Flags()
	if _, ok := tool.(tools.RunsInDirectory); ok {
		flags.StringP("directory", "d", "", "Directory to scan")
	}
	if _, ok := tool.(tools.RunsWithAPIClient); !ok {
		opts.AuthNotRequired = true
	}
	return c
}

func meta(cmds []*cobra.Command) *cobra.Command {
	c := &cobra.Command{
		Use:   "all",
		Short: "run all IaC scanning tools",
		Args:  cobra.NoArgs,
	}
	for i := range cmds {
		c.Flags().AddFlagSet(cmds[i].NonInheritedFlags())
	}
	// TODO: for now, we run all tools unconditionally.
	c.Flags().Bool("no-fail", true, "Ignore failures in individual tools, continuing with the remaining tools")

	c.RunE = func(cmd *cobra.Command, args []string) error {
		cmds := cmds
		for i := range cmds {
			// prevent recursion
			if cmds[i].Use == c.Use {
				continue
			}
			if err := cmds[i].RunE(cmd, args); err != nil {
				if nofail, _ := cmd.Flags().GetBool("no-fail"); !nofail {
					return err
				}
			}
		}
		return nil
	}
	return c
}

func cloudformationGuard() *cobra.Command {
	t := &cloudformationguard.Tool{}
	c := createCommand(t)
	flags := c.Flags()
	flags.StringVar(&t.File, "file", "", "The cloudformation file to scan")
	_ = c.MarkFlagRequired("file")
	return c
}

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "iac-scan",
		Short: "Run an Infrastructure-as-code scanner",
		Example: `  # run the default scanner in the current directory
  iac-scan default -d .`,
	}
	t := createCommand(&terrascan.Tool{})
	t.Aliases = []string{"default"}
	c.AddCommand(t)
	c.AddCommand(createCommand(&checkov.Tool{}))
	c.AddCommand(createCommand(&tfsec.Tool{}))
	c.AddCommand(createCommand(&cfnpythonlint.Tool{}))
	c.AddCommand(meta(c.Commands()))

	// cloudformationguard is not supported by the meta/all scanner, at least for now.
	// it is the only weird tool that requires explicit file input.
	c.AddCommand(cloudformationGuard())
	return c
}
