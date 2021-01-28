package cloudscan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools/cloudsploit"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "cloud-scan",
		Short: "Scan cloud infrastructure",
		Args:  cobra.NoArgs,
	}
	c.AddCommand(cloudsploit.Command())
	return c
}
