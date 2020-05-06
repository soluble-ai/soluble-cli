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

package aws

import (
	"fmt"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "aws",
		Short: "Manage AWS integration",
	}
	c.AddCommand(setupCommand())
	return c
}

func setupCommand() *cobra.Command {
	var opts options.ClientOpts
	var awsAccount string
	c := &cobra.Command{
		Use:   "get-setup",
		Short: "Return the aws CLI commands to create a role for Soluble to use",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient := opts.GetAPIClient()
			response, err := apiClient.GetClient().R().
				SetQueryParam("awsAccount", awsAccount).
				SetHeader("Accept", "text/plain").
				Get("org/{org}/aws/iam-setup-script")
			if err != nil {
				return err
			}
			fmt.Println(string(response.Body()))
			return nil
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			log.Level = log.Error
		},
	}
	opts.Register(c)
	c.Flags().StringVar(&awsAccount, "aws-account", "", "The AWS account number (required)")
	_ = c.MarkFlagRequired("aws-account")
	return c
}
