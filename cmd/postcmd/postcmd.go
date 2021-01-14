package postcmd

import (
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var (
		module string
		files  []string
		values map[string]string
	)
	opts := options.PrintClientOpts{
		PrintOpts: options.PrintOpts{
			DefaultOutputFormat: "value(assessment.appUrl)",
		},
	}
	c := &cobra.Command{
		Use:   "post",
		Short: "Send data to soluble",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := opts.GetAPIClient().XCPPost(opts.GetOrganization(), module, files, values,
				xcp.WithCIEnv(""))
			if err != nil {
				return err
			}
			opts.PrintResult(result)
			return nil
		},
	}
	opts.Register(c)
	flags := c.Flags()
	// on unconditionally -- hidden for backwards compatibility
	flags.BoolP("env", "e", false, "Include CI environment variables and git information (always enabled).")
	_ = flags.MarkHidden("env")
	flags.StringVarP(&module, "module", "m", "", "The module to post under, required.")
	flags.StringSliceVarP(&files, "file", "f", nil, "Send a file, can be repeated")
	flags.StringToStringVarP(&values, "param", "p", nil, "Include a key value pair, can be repeated.  The argument should be in the form key=value.")
	_ = c.MarkFlagRequired("module")
	return c
}
