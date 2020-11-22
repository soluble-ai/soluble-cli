package tools

import (
	"fmt"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/blurb"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/version"
	"github.com/spf13/cobra"
)

func CreateCommand(tool Interface) *cobra.Command {
	c := &cobra.Command{
		Use:  tool.Name(),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTool(tool)
		},
	}
	tool.Register(c)
	return c
}

func runTool(tool Interface) error {
	opts := tool.GetToolOptions()
	if opts.UploadEnabled && opts.GetAPIClientConfig().APIToken == "" {
		blurb.SignupBlurb(opts, "{info:--upload} requires signing up with {primary:Soluble}.", "")
		return fmt.Errorf("not authenticated with Soluble")
	}
	result, err := tool.Run()
	if err != nil {
		return err
	}
	result.AddValue("TOOL_NAME", tool.Name()).
		AddValue("CLI_VERSION", version.Version)
	if result.Data != nil && result.PrintPath != nil {
		// include the print config in the results
		p := result.Data.PutObject("soluble_print_config")
		p.Put("print_path", jnode.FromSlice(result.PrintPath))
		p.Put("print_columns", jnode.FromSlice(result.PrintColumns))
	}

	assessmentURL := ""
	if opts.UploadEnabled {
		response, err := result.report(tool)
		if err != nil {
			return err
		}
		assessmentURL = response.Path("assessment").Path("appUrl").AsText()
	}
	opts.Path = result.PrintPath
	opts.Columns = result.PrintColumns
	opts.PrintResult(result.Data)
	if !opts.UploadEnabled {
		blurb.SignupBlurb(opts, "Want to manage results centrally with {primary:Soluble}?", "run this command again with the {info:--upload} flag")
	}
	if assessmentURL != "" {
		log.Infof("Results uploaded, see {primary:%s} for more information", assessmentURL)
	}

	return nil
}
