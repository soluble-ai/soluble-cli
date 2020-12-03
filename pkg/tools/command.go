package tools

import (
	"fmt"

	"github.com/soluble-ai/soluble-cli/pkg/blurb"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/spf13/cobra"
)

func CreateCommand(tool Interface, c *cobra.Command) *cobra.Command {
	if c == nil {
		c = &cobra.Command{
			Use:  tool.Name(),
			Args: cobra.NoArgs,
		}
	}
	c.RunE = func(cmd *cobra.Command, args []string) error {
		return runTool(tool)
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
	result, err := opts.RunTool(tool)
	if err != nil {
		return err
	}
	opts.Path = result.PrintPath
	opts.Columns = result.PrintColumns
	opts.PrintResult(result.Data)
	if !opts.UploadEnabled {
		blurb.SignupBlurb(opts, "Want to manage findings with {primary:Soluble}?", "run this command again with the {info:--upload} flag")
	}
	if result.AssessmentURL != "" {
		log.Infof("Results uploaded, see {primary:%s} for more information", result.AssessmentURL)
	}
	return nil
}
