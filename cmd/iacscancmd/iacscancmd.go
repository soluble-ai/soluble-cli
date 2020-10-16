package iacscancmd

import (
	"github.com/soluble-ai/soluble-cli/pkg/iacscan"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var dir string
	opts := options.PrintOpts{}
	c := &cobra.Command{
		Use:   "iac-scan",
		Short: "Run an Infrastructure-as-code scanner",
		RunE: func(cmd *cobra.Command, args []string) error {
			scanner := iacscan.New(&iacscan.StockTerrascan{
				Directory: dir,
			})
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
