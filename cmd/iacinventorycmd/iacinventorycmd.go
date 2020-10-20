package iacinventorycmd

import (
	"github.com/soluble-ai/soluble-cli/pkg/iacinventory"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var username string
	var oauthToken string
	opts := options.PrintOpts{}
	c := &cobra.Command{
		Use:   "iac-inventory",
		Short: "Run an Infrastructure-as-code inventorier on repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			// For now, we only support Github.
			scanner := iacinventory.New(&iacinventory.GithubScanner{
				User:       username,
				OauthToken: oauthToken,
			})
			result, err := scanner.Run()
			if err != nil {
				return err
			}
			opts.PrintResult(result)
			return nil
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVar(&username, "gh-username", "", "Github Username")
	flags.StringVar(&oauthToken, "gh-oauthtoken", "", "Github OAuthToken")
	return c
}
