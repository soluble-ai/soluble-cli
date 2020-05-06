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
	"github.com/soluble-ai/soluble-cli/pkg/client"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/spf13/cobra"
)

type HasClientOpts interface {
	GetClientOpts() *ClientOpts
}

type ClientOpts struct {
	client.Config
	AuthNotRequired bool
}

var _ HasClientOpts = &ClientOpts{}

func (opts *ClientOpts) GetClientOpts() *ClientOpts {
	return opts
}

func (opts *ClientOpts) Register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&opts.APIServer, "api-server", "", "Soluble API server endpoint (e.g. https://api.soluble.cloud)")
	cmd.Flags().BoolVarP(&opts.TLSNoVerify, "disable-tls-verify", "k", false, "Disable TLS verification on api-server")
	if !opts.AuthNotRequired {
		cmd.Flags().StringVar(&opts.Organization, "organization", "", "The organization to use.")
		cmd.Flags().StringVar(&opts.APIToken, "token", "", "The authentication token (read from profile by default)")
	}
}

func (opts *ClientOpts) GetAPIClientConfig() *client.Config {
	cfg := opts.Config
	if cfg.Organization == "" {
		cfg.Organization = config.Config.Organization
	}
	if cfg.APIToken == "" {
		cfg.APIToken = config.Config.APIToken
	}
	if cfg.APIServer == "" {
		cfg.APIServer = config.Config.APIServer
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

func (opts *ClientOpts) GetAPIClient() client.Interface {
	return client.NewClient(opts.GetAPIClientConfig())
}

func (opts *ClientOpts) GetUnauthenticatedAPIClient() client.Interface {
	cfg := opts.GetAPIClientConfig()
	cfg.APIToken = ""
	return client.NewClient(cfg)
}
