package tools

import (
	"fmt"

	"github.com/soluble-ai/soluble-cli/pkg/blurb"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/spf13/cobra"
)

type HasCommandTemplate interface {
	CommandTemplate() *cobra.Command
}

func CreateCommand(tool Interface) *cobra.Command {
	var c *cobra.Command
	if ct, ok := tool.(HasCommandTemplate); ok {
		c = ct.CommandTemplate()
		if c.Args == nil {
			c.Args = cobra.NoArgs
		}
	} else {
		c = &cobra.Command{
			Use:   tool.Name(),
			Short: fmt.Sprintf("Run %s", tool.Name()),
			Args:  cobra.NoArgs,
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
	if err := opts.Validate(); err != nil {
		return err
	}
	result, err := opts.RunTool(tool, true)
	if err != nil || result == nil {
		return err
	}
	if !opts.UploadEnabled {
		blurb.SignupBlurb(opts, "Want to manage findings with {primary:Soluble}?", "run this command again with the {info:--upload} flag")
	}
	if result.Assessment != nil && result.Assessment.URL != "" {
		log.Infof("Results uploaded, see {primary:%s} for more information", result.Assessment.URL)
	}
	return nil
}
