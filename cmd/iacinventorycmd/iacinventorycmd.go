package iacinventorycmd

import (
	"github.com/soluble-ai/soluble-cli/pkg/iacinventory"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var dir string
	opts := options.PrintOpts{}
	c := &cobra.Command{
		Use:   "iac-inventory",
		Short: "Run an Infrastructure-as-code inventorier on repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			// For now, we only support Github.
			scanner := iacinventory.New(&iacinventory.GithubScanner{})
			result, err := scanner.Run()
			if err != nil {
				return err
			}
			opts.PrintResult(result)
			return nil
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVarP(&dir, "directory", "d", "", "Directory to scan")
	_ = c.MarkFlagRequired("directory")
	return c
}
