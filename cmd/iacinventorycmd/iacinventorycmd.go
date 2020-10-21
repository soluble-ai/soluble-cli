package iacinventorycmd

import (
	"bytes"
	"encoding/json"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/iacinventory"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var username string
	var oauthToken string
	var submit bool
	opts := options.PrintClientOpts{}
	c := &cobra.Command{
		Use:   "iac-inventory",
		Short: "Run an Infrastructure-as-code inventorier on repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			// For now, we only support Github.
			inventorier := iacinventory.New(&iacinventory.GithubInventorier{
				User:       username,
				OauthToken: oauthToken,
			})
			result, err := inventorier.Run()
			if err != nil {
				return err
			}
			n := jnode.FromSlice(result)
			opts.PrintResult(n)
			j, err := json.MarshalIndent(result, "", "    ")
			if err != nil {
				return err
			}
			if !submit {
				// TODO: also early exit if auth is not configured
				return nil
			}

			values := make(map[string]string) // TODO: add debugging values?
			c := opts.GetAPIClient().GetClient()
			req := c.R()
			req.SetFileReader("file_iac_inventory", "iac_inventory.json", bytes.NewReader(j))
			req.SetHeader("X-SOLUBLE-ORG-ID", opts.GetAPIClientConfig().Organization)
			req.SetMultipartFormData(values)
			_, err = req.Post("/api/v1/xcp/iac-inventory/data")
			return err
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVar(&username, "gh-username", "", "Github Username")
	flags.StringVar(&oauthToken, "gh-oauthtoken", "", "Github OAuthToken")
	flags.BoolVar(&submit, "submit", false, "submit results to the Soluble API")
	return c
}
