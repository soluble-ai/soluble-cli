package iacscancmd

import (
	"github.com/manifoldco/promptui"
	"github.com/soluble-ai/soluble-cli/pkg/iacscan"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var dir string
	var report bool
	var checkTool string

	opts := options.PrintOpts{}
	c := &cobra.Command{
		Use:   "iac-scan",
		Short: "Run an Infrastructure-as-code scanner",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(checkTool) == 0 {
				prompt := promptui.Select{
					Label: "Select Tool",
					Items: []string{"terrascan", "kube-bench", "tfsec", "kube-score", "kube-audit", "cfn-guard"},
				}

				_, tool, err := prompt.Run()
				if err != nil {
					return err
				}
				log.Infof("You choose %q", tool)
				checkTool = tool
			}
			scanner := iacscan.New(&iacscan.StockTerrascan{
				Directory: dir,
				Report:    report,
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
	flags.StringVarP(&checkTool, "check-tool", "t", "", "Tool to use for Infrastructure as Code scanner")
	flags.BoolVarP(&report, "report", "r", true, "Report back to control plane")
	_ = c.MarkFlagRequired("directory")
	return c
}
