package iacscan

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/terrascan"
	"github.com/soluble-ai/soluble-cli/pkg/tools/tfsec"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var (
		directory   string
		scannerType string
	)
	opts := &tools.ToolOpts{}
	c := &cobra.Command{
		Use:   "iac-scan",
		Short: "Run an Infrastructure-as-code scanner",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if scannerType == "" {
				prompt := promptui.Select{
					Label: "Select Tool",
					Items: []string{"terrascan", "tfsec"},
				}
				_, selection, err := prompt.Run()
				if err != nil {
					return err
				}
				scannerType = selection
			}
			var tool tools.Interface
			switch scannerType {
			case "terrascan":
				tool = &terrascan.Tool{
					APIClient: opts.GetAPIClient(),
					Directory: directory,
				}
			case "tfsec":
				tool = &tfsec.Tool{
					Directory: directory,
				}
			default:
				return fmt.Errorf("unknown scanner type %s", scannerType)
			}
			return opts.RunTool(tool)
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVarP(&directory, "directory", "d", "", "Directory to scan")
	flags.StringVar(&scannerType, "scanner-type", "", "The scanner to use")
	_ = c.MarkFlagRequired("directory")
	return c
}
