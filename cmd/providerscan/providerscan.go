package providerscan

import "github.com/spf13/cobra"

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "provider-scan",
		Short: "Scan cloud infrastructure providers",
		Long: `Scan cloud infrastructure providers (AWS, GCP, Azure).

Example: scan AWS with cloudsploit:
  $ solube provider-scan cloudsploit aws`,
		Args: cobra.NoArgs,
	}
	c.AddCommand(cloudsploitCommand())
	return c
}
