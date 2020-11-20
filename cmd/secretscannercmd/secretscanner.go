package secretscanner

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/secrets"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return tools.CreateCommand(&secrets.Tool{},
		&cobra.Command{
			Use:   "secrets-scan",
			Short: "Scans Git repositories for any sensitive information such as private keys, API secrets and tokens, etc",
			Args:  cobra.NoArgs,
		})
}
