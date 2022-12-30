// Copyright 2020 Soluble Inc
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

package options

import (
	"fmt"
	"time"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ClientOpts struct {
	APIConfig      api.Config
	DefaultTimeout int

	nonComponentOrgFlag string
	client              *api.Client
	unauthClient        *api.Client
}

var _ Interface = &ClientOpts{}

func GetClientOptionsGroupHelpCommand() *cobra.Command {
	opts := &ClientOpts{}
	return opts.GetClientOptionsGroup().GetHelpCommand()
}

func (opts *ClientOpts) SetContextValues(context map[string]string) {
	context["organizationID"] = opts.APIConfig.Organization
}

func (opts *ClientOpts) GetClientOptionsGroup() *HiddenOptionsGroup {
	return &HiddenOptionsGroup{
		Name: "client-options",
		Long: "These flags control how the CLI connects to lacework IAC",
		CreateFlagsFunc: func(flags *pflag.FlagSet) {
			flags.StringVar(&opts.APIConfig.APIServer, "api-server", "", "The lacework IAC API server `url` (by default $SOLUBLE_API_URL if set, or https://api.soluble.cloud)")
			flags.BoolVarP(&opts.APIConfig.TLSNoVerify, "disable-tls-verify", "k", false, "Disable TLS verification on api-server")
			flags.DurationVar(&opts.APIConfig.Timeout, "api-timeout", time.Duration(opts.DefaultTimeout)*time.Second,
				"The `timeout` (e.g. 15s, 500ms) for API requests (0 means no timeout)")
			flags.IntVar(&opts.APIConfig.RetryCount, "api-retry", 5, "The `number` of times to retry the request")
			flags.Float64Var(&opts.APIConfig.RetryWaitSeconds, "api-retry-wait", 1,
				"The initial time in `seconds` to wait between retry attempts, e.g. 0.5 to wait 500 millis")
			flags.StringSliceVar(&opts.APIConfig.Headers, "api-header", nil, "Set custom headers in the form `name:value` on requests")
			flags.StringVar(&opts.APIConfig.Organization, "iac-organization", "", "The IAC organization `id` to use (by default $LW_IAC_ORGANIZATION if set.)")
			if !config.IsRunningAsComponent() {
				// The lacework CLI eats --organization, so we can only define this
				// flag when not running as a component.
				flags.StringVar(&opts.nonComponentOrgFlag, "organization", "",
					"The soluble organization `id` to use.  Overrides the value of --iac-organization.")
			}
			flags.StringVar(&opts.APIConfig.LegacyAPIToken, "iac-api-token", "", "The legacy authentication `token` (read from profile by default)")
			flags.StringVar(&opts.APIConfig.LaceworkAccount, "account", "", "The Lacework account")
		},
	}
}

func (opts *ClientOpts) Register(cmd *cobra.Command) {
	opts.GetClientOptionsGroup().Register(cmd)
}

func (opts *ClientOpts) MustGetAPIClient() *api.Client {
	c, err := opts.GetAPIClient()
	if err != nil {
		panic(err)
	}
	return c
}

func (opts *ClientOpts) GetAPIClient() (*api.Client, error) {
	if opts.client == nil {
		if opts.nonComponentOrgFlag != "" {
			opts.APIConfig.Organization = opts.nonComponentOrgFlag
		}
		opts.APIConfig.SetValues()
		err := opts.APIConfig.Validate()
		if err != nil {
			return nil, err
		}
		opts.client = api.NewClient(&opts.APIConfig)
		opts.client.NoOrganizationHook = func() error {
			if err := opts.ConfigureDefaultOrganization(); err != nil {
				return err
			}
			return config.Save()
		}
	}
	return opts.client, nil
}

func (opts *ClientOpts) GetUnauthenticatedAPIClient() *api.Client {
	if opts.unauthClient == nil {
		cfg := opts.APIConfig
		cfg.SetValues()
		cfg.LegacyAPIToken = ""
		cfg.LaceworkAPIToken = ""
		cfg.LaceworkAccount = ""
		opts.unauthClient = api.NewClient(&cfg)
	}
	return opts.unauthClient
}

func (opts *ClientOpts) IsAuthenticated() bool {
	opts.APIConfig.SetValues()
	if opts.APIConfig.LegacyAPIToken != "" {
		return true
	}
	return opts.APIConfig.LaceworkAPIKey != "" && opts.APIConfig.LaceworkAPISecret != ""
}

func (opts *ClientOpts) RequireAuthentication() error {
	if !opts.IsAuthenticated() {
		log.Warnf("This command requires signing up with {primary:Lacework}")
		return fmt.Errorf("not authenticated with Lacework")
	}
	return nil
}

func (opts *ClientOpts) ConfigureDefaultOrganization() error {
	n, err := opts.client.Get("/api/v1/users/profile")
	if err != nil {
		return fmt.Errorf("could not get IAC organization - %w", err)
	}
	log.Debugf("%s", n)
	orgID := getText(n.Path("data").Path("defaultOrgId"))
	if orgID == "" {
		orgID = getText(n.Path("data").Path("currentOrgId"))
	}
	if orgID == "" {
		orgID = getText(n.Path("data").Path("organizations").Get(0).Path("orgId"))
	}
	if orgID == "" {
		return fmt.Errorf("could not determine default IAC organization")
	}
	log.Infof("Configuring IAC to use organization {primary:%s}", orgID)
	config.Get().Organization = orgID
	config.Get().ConfiguredAccount = opts.client.LaceworkAccount
	opts.client.Organization = orgID
	return nil
}

func (opts *ClientOpts) ValidateOrganization() error {
	n, err := opts.client.Get("/api/v1/users/profile")
	if err != nil {
		return fmt.Errorf("could not get IAC organizations - %w", err)
	}
	for _, org := range n.Path("data").Path("organizations").Elements() {
		orgID := getText(org.Path("orgId"))
		if orgID == opts.client.Organization {
			return nil
		}
	}
	return fmt.Errorf("invalid IAC organization")
}

func getText(n *jnode.Node) string {
	if n.GetType() == jnode.Text || n.GetType() == jnode.Number {
		return n.AsText()
	}
	return ""
}
