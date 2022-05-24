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
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ClientOpts struct {
	api.Config
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
	context["organizationID"] = opts.GetOrganization()
}

func (opts *ClientOpts) GetClientOptionsGroup() *HiddenOptionsGroup {
	return &HiddenOptionsGroup{
		Name: "client-options",
		Long: "These flags control how the CLI connects to Soluble",
		CreateFlagsFunc: func(flags *pflag.FlagSet) {
			flags.StringVar(&opts.APIServer, "api-server", "", "Soluble API server `url` (e.g. https://api.soluble.cloud)")
			flags.BoolVarP(&opts.TLSNoVerify, "disable-tls-verify", "k", false, "Disable TLS verification on api-server")
			flags.DurationVar(&opts.Timeout, "api-timeout", time.Duration(opts.DefaultTimeout)*time.Second,
				"The `timeout` (e.g. 15s, 500ms) for API requests (0 means no timeout)")
			flags.IntVar(&opts.RetryCount, "api-retry", 0, "The `number` of times to retry the request")
			flags.Float64Var(&opts.RetryWaitSeconds, "api-retry-wait", 0,
				"The initial time in `seconds` to wait between retry attempts, e.g. 0.5 to wait 500 millis")
			flags.StringSliceVar(&opts.Headers, "api-header", nil, "Set custom headers in the form `name:value` on requests")
			flags.StringVar(&opts.Organization, "organization", "", "The organization `id` to use.")
			flags.StringVar(&opts.APIToken, "api-token", "", "The authentication `token` (read from profile by default)")
		},
	}
}

func (opts *ClientOpts) Register(cmd *cobra.Command) {
	opts.GetClientOptionsGroup().Register(cmd)
}

func (opts *ClientOpts) GetAPIClientConfig() *api.Config {
	cfg := opts.Config

	if cfg.Organization == "" {
		cfg.Organization = config.Config.Organization
	}
	if cfg.APIToken == "" {
		cfg.APIToken = config.Config.GetAPIToken()
	}
	if cfg.APIServer == "" {
		cfg.APIServer = config.Config.GetAPIServer()
	}
	if cfg.APIServer == "" {
		cfg.APIServer = "https://api.soluble.cloud"
	}
	if !cfg.TLSNoVerify {
		cfg.TLSNoVerify = config.Config.TLSNoVerify
	}
	return &cfg
}

func (opts *ClientOpts) GetOrganization() string {
	if opts.Organization != "" {
		return opts.Organization
	}
	return config.Config.Organization
}

func (opts *ClientOpts) GetAPIClient() *api.Client {
	if opts.client == nil {
		opts.client = api.NewClient(opts.GetAPIClientConfig())
	}
	return opts.client
}

func (opts *ClientOpts) GetUnauthenticatedAPIClient() *api.Client {
	if opts.unauthClient == nil {
		cfg := opts.GetAPIClientConfig()
		cfg.APIToken = ""
		opts.unauthClient = api.NewClient(cfg)
	}
	return opts.unauthClient
}

func (opts *ClientOpts) IsAuthenticated() bool {
	return opts.GetAPIClientConfig().APIToken != ""
}

func (opts *ClientOpts) RequireAPIToken() error {
	if opts.GetAPIClientConfig().APIToken == "" {
		SignupBlurb(opts, "This command requires signing up with {primary:Soluble} (unless --upload=false).", "")
		return fmt.Errorf("not authenticated with Soluble")
	}
	return nil
}
