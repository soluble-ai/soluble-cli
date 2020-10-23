package iacscancmd

import (
	"github.com/soluble-ai/soluble-cli/pkg/iacscan"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	config := iacscan.Config{}
	opts := options.PrintClientOpts{}
	c := &cobra.Command{
		Use:   "iac-scan",
		Short: "Run an Infrastructure-as-code scanner",
		RunE: func(cmd *cobra.Command, args []string) error {
			config.APIClient = opts.GetAPIClient()
			config.Organizaton = opts.GetOrganization()
			scanner, err := iacscan.New(config)
			if err != nil {
				return err
			}
			result, err := scanner.Run()
			if err != nil {
				return err
			}
			opts.Path = result.PrintPath
			opts.Columns = result.PrintColumns
			opts.PrintResult(result.N)
			return nil
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVarP(&config.Directory, "directory", "d", "", "Directory to scan")
	flags.BoolVarP(&config.ReportEnabled, "report", "r", false, "Upload scan results to soluble")
	flags.StringVar(&config.ScannerType, "scanner-type", "terrascan", "The scanner to use")
	_ = c.MarkFlagRequired("directory")
	return c
}
