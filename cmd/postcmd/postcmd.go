package postcmd

import (
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var (
		module string
		files  []string
		values map[string]string
	)
	opts := options.ClientOpts{}
	c := &cobra.Command{
		Use:   "post",
		Short: "Send data to soluble",
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.GetAPIClient().XCPPost(opts.GetOrganization(), module, files, values)
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVarP(&module, "module", "m", "", "The module to post under, required.")
	flags.StringSliceVarP(&files, "file", "f", nil, "Send a file, can be repeated")
	flags.StringToStringVarP(&values, "param", "p", nil, "Include a key value pair, can be repeated.  The argument should be in the form key=value.")
	_ = c.MarkFlagRequired("module")
	return c
}
