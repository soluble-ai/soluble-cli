package semgrep

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/semgrep"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "semgrep",
		Short: "Run semgrep",
	}
	return tools.CreateCommand(semgrep.Tool{}, c)
}
