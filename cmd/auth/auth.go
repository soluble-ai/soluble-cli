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

package auth

import (
	"fmt"

	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}
	c.AddCommand(profileCmd())
	if !config.IsRunningAsComponent() {
		c.AddCommand(
			printTokenCmd(),
			setAccessTokenCmd(),
		)
	}
	return c
}

func profileCmd() *cobra.Command {
	opts := options.PrintClientOpts{}
	c := &cobra.Command{
		Use:   "profile",
		Short: "Display the user's IAC profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := opts.GetAPIClient()
			if err != nil {
				return err
			}
			result, err := apiClient.Get("/api/v1/users/profile")
			if err != nil {
				return err
			}
			opts.PrintResult(result)
			return nil
		},
	}
	opts.Register(c)
	return c
}

func printTokenCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "print-access-token",
		Short: "Print the access token (for use with curl)",
		Long: `Print the access token, e.g. for use with curl:
		
curl -H “Authorization: Bearer $(soluble print-access-token)” ...`,
		Deprecated: "use 'lacework access-token' instead",
		Args:       cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.Config.GetAPIToken() == "" {
				return fmt.Errorf("not authenticated, use login to authenticate")
			}
			fmt.Println(config.Config.GetAPIToken())
			return nil
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			log.Level = log.Error
		},
	}
	return c
}

func setAccessTokenCmd() *cobra.Command {
	opts := options.PrintClientOpts{}
	var accessToken string
	c := &cobra.Command{
		Use:         "set-access-token",
		Short:       "Add an access token",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{config.ConfigurationNotRequired: "1"},
		RunE: func(cmd *cobra.Command, args []string) error {
			config.Config.APIToken = accessToken
			cfg, err := opts.GetAPIClientConfig()
			if err != nil {
				return err
			}
			log.Infof("Verifying access token with {primary:%s}", cfg.APIServer)
			apiClient, err := opts.GetAPIClient()
			if err != nil {
				return err
			}
			_, err = apiClient.Get("/api/v1/users/profile")
			if err != nil {
				return err
			}
			config.Config.APIServer = cfg.APIServer
			config.Config.TLSNoVerify = cfg.TLSNoVerify
			if err := config.Save(); err != nil {
				return err
			}
			log.Infof("Current org is {info:%s}", config.Config.Organization)
			return nil
		},
	}
	opts.Register(c)
	c.Flags().StringVar(&accessToken, "access-token", "", "The legacy IAC access `token`")
	_ = c.MarkFlagRequired("access-token")
	return c
}
