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

var supportedTools = []tools.Interface{
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
	toolNames := []string{}
	for _, tool := range supportedTools {
		toolNames = append(toolNames, tool.Name())
	}
	c := &cobra.Command{
		Use:   "iac-scan",
		Short: "Run an Infrastructure-as-code scanner",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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
			var tool tools.Interface
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
			if u, ok := tool.(tools.RunsInDirectory); ok {
				if directory == "" {
					return fmt.Errorf("the %s scanner requires a directory to run in", scannerType)
				}
				u.SetDirectory(directory)
			}
			if u, ok := tool.(tools.RunsWithAPIClient); ok {
				u.SetAPIClient(opts.GetAPIClient())
			}
			return opts.RunTool(tool)
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVarP(&directory, "directory", "d", "", "Directory to scan")
	flags.StringVar(&scannerType, "scanner-type", "",
		fmt.Sprintf("The scanner to use, should be one of: %s", strings.Join(toolNames, " ")))
	return c
}
