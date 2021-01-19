package codescan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/semgrep"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&semgrep.Tool{})
	c.Use = "code-scan"
	c.Short = "Scan code with semgrep"
	return c
}
