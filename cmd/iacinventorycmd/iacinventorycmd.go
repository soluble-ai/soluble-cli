package iacinventorycmd

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/iacinventory"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	tool := &iacinventory.GithubIacInventoryScanner{}
	opts := tools.ToolOpts{}
	c := &cobra.Command{
		Use:   "iac-inventory",
		Short: "Look for infrastructure-as-code and optionally send the inventory to Soluble",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.RunTool(tool)
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVar(&tool.User, "gh-username", "", "Github Username")
	flags.StringVar(&tool.OauthToken, "gh-oauthtoken", "", "Github OAuthToken")
	flags.BoolVar(&tool.AllRepos, "all", false, "Inventory all accessible public and private repostories.")
	flags.BoolVar(&tool.PublicRepos, "public", false, "Inventory accessible public repositories.")
	flags.StringSliceVar(&tool.ExplicitRepositories, "repository", nil, "Inventory this repository.  May be repeated.")
	return c
}
