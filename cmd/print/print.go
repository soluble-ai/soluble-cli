package print

import (
	"io"
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var (
		opts options.PrintOpts
	)
	c := &cobra.Command{
		Use:   "print file",
		Short: "Print a JSON document",
		Long: `Print a JSON document with the common printing options.

The JSON document will be read from "file".  Use "-" to read from stdin.

See help-print-options for more details.

This command can avoid having to repeatedly run assessments to develop
print formats.  For example:

; soluble tf-scan -d ~/my-work --format json > assesments.json
; soluble print --print-template '{{ range (index . 0).findings }}{{ printf "%s %s\n" .sid .severity }}{{ end }}' assessments.json
ckv-aws-24 Critical
ckv-aws-25 Critical
ckv-aws-23 Low

The table and csv printers must be supplied a JSON array to print, and
each element in that array should be an object.  Use the --path flag to
give the simple path to the array.  A quick example:

; echo '{ "results": [ { "X": 1, "Y": 2  } ] }' | soluble print --path results --columns X,Y -
X    Y
1    2

`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file := args[0]
			var (
				dat []byte
				err error
			)
			if file == "-" {
				dat, err = io.ReadAll(os.Stdin)
			} else {
				dat, err = os.ReadFile(file)
			}
			if err != nil {
				return err
			}
			n, err := jnode.FromJSON(dat)
			if err != nil {
				return err
			}
			opts.PrintResult(n)
			return nil
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringSliceVar(&opts.Columns, "columns", nil,
		"Configure the printed columns for the table format.  May be repeated.")
	flags.StringSliceVar(&opts.Path, "path", nil, "For table print, print the data at `path`.")
	return c
}
