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

	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ClientOpts struct {
	APIConfig      api.Config
	DefaultTimeout int

	client       *api.Client
	unauthClient *api.Client
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
			flags.IntVar(&opts.APIConfig.RetryCount, "api-retry", 0, "The `number` of times to retry the request")
			flags.Float64Var(&opts.APIConfig.RetryWaitSeconds, "api-retry-wait", 0,
				"The initial time in `seconds` to wait between retry attempts, e.g. 0.5 to wait 500 millis")
			flags.StringSliceVar(&opts.APIConfig.Headers, "api-header", nil, "Set custom headers in the form `name:value` on requests")
			flags.StringVar(&opts.APIConfig.Organization, "iac-organization", "", "The IAC organization `id` to use (by default $LW_IAC_ORGANIZATION if set.)")
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
		opts.APIConfig.SetValues()
		err := opts.APIConfig.Validate(true)
		if err != nil {
			return nil, err
		}
		opts.client = api.NewClient(&opts.APIConfig)
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
		log.Warnf("This command requires signing up with {primary:Lacework} (unless --upload=false).")
		return fmt.Errorf("not authenticated with Lacework")
	}
	return nil
}
