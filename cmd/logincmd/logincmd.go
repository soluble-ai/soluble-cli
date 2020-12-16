package logincmd

import (
	"fmt"

	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/login"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	opts := options.PrintOpts{}
	var (
		app      string
		reset    bool
		headless bool
	)
	c := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with soluble",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.Config.APIToken != "" && !reset {
				log.Infof("Already logged in to {primary:%s}, use --reset to re-authenticate", config.Config.APIServer)
				return nil
			}
			if app == "" {
				app = config.Config.GetAppURL()
			}
			flow := login.NewFlow(app, headless)
			resp, err := flow.Run()
			if err != nil {
				log.Errorf("Authentication did not complete: {danger:%s}", err)
				log.Infof("See {primary:https://docs.soluble.ai/cli/} for more information")
				return fmt.Errorf("failed")
			}
			config.Config.APIServer = resp.APIServer
			config.Config.APIToken = resp.Token
			config.Config.Organization = resp.OrgID
			config.Save()
			log.Infof("Authentication successful")
			return nil
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVar(&app, "app", "", "The app URL to authenticate with")
	flags.BoolVar(&reset, "reset", false, "Re-authenticate, even if an auth token is already present")
	flags.BoolVar(&headless, "headless", false, "Don't try and open a browser to complete the flow")
	_ = flags.MarkHidden("app")
	return c
}
