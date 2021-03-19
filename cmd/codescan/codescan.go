package codescan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/bandit"
	"github.com/soluble-ai/soluble-cli/pkg/tools/semgrep"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "code-scan",
		Short: "Scan code with a variety of static analysis tools",
	}
	c.AddCommand(
		tools.CreateCommand(&semgrep.Tool{}),
		tools.CreateCommand(&bandit.Tool{}),
	)
	return c
}
