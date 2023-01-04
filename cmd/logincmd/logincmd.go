// Copyright 2021 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	opts := options.PrintClientOpts{}
	var (
		app      string
		reset    bool
		headless bool
	)
	c := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Soluble",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.Get().APIToken != "" && !reset {
				log.Infof("Already logged in to {primary:%s}, use --reset to re-authenticate", config.Get().APIServer)
				return nil
			}
			if reset {
				config.Get().Reset()
			}
			if app == "" {
				cfg := opts.APIConfig.SetValues()
				app = cfg.GetAppURL()
			}
			flow := login.NewFlow(app, headless)
			resp, err := flow.Run()
			if err != nil {
				log.Errorf("Authentication did not complete: {danger:%s}", err)
				log.Infof("See {primary:https://github.com/soluble-ai/soluble-cli} for more information")
				return fmt.Errorf("failed")
			}
			config.Get().APIServer = resp.APIServer
			config.Get().APIToken = resp.Token
			config.Get().Organization = resp.OrgID
			defer log.Infof("Authentication successful, IAC organization is {info:%s}", resp.OrgID)
			return config.Save()
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVar(&app, "app", "", "The app URL to authenticate with")
	flags.BoolVar(&reset, "reset", false, "Re-authenticate, even if an auth token is already present")
	flags.BoolVar(&headless, "headless", false, "Don't try and open a browser to complete the flow")
	_ = flags.MarkHidden("app")
	if config.IsRunningAsComponent() {
		c.Hidden = true
	}
	return c
}
