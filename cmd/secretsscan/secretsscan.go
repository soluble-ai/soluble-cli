package secretsscan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/secrets"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&secrets.Tool{})
	c.Use = "secrets-scan"
	c.Short = "Scan for secrets in code"
	return c
}
