package iacscan

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/soluble-ai/soluble-cli/pkg/tools/terrascan"
	"github.com/soluble-ai/soluble-cli/pkg/tools/tfsec"
	"github.com/spf13/cobra"
)

var supportedTools = []tools.InterfaceWithDirectory{
	&terrascan.Tool{},
	&tfsec.Tool{},
	&checkov.Tool{},
}

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
			toolNames := []string{}
			for _, tool := range supportedTools {
				toolNames = append(toolNames, tool.Name())
			}

			if scannerType == "" {
				prompt := promptui.Select{
					Label: "Select Tool",
					Items: toolNames,
				}
				_, selection, err := prompt.Run()
				if err != nil {
					return err
				}
				scannerType = selection
			}
			var tool tools.InterfaceWithDirectory
			for _, t := range supportedTools {
				if t.Name() == scannerType {
					tool = t
					break
				}
			}
			if tool == nil {
				return fmt.Errorf("unknown scanner type %s, must be one of: %s", scannerType,
					strings.Join(toolNames, " "))
			}
			tool.SetDirectory(directory)
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
