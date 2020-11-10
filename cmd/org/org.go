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

package org

import (
	"io/ioutil"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "org",
		Short: "Manage Organization",
	}
	c.AddCommand(setConfigCommand())
	return c
}

func setConfigCommand() *cobra.Command {
	var file string
	opts := options.PrintClientOpts{}
	c := &cobra.Command{
		Use:   "set-config",
		Short: "Sets the config for user's current organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}

			body, err := jnode.FromJSON(content)
			if err != nil {
				return err
			}

			result, err := opts.GetAPIClient().Post("/api/v1/org/{org}/config", body)
			if err != nil {
				return err
			}
			log.Infof("Current org's config set to {info:%s}", result)
			return nil
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVarP(&file, "file", "f", "", "Send a config file (json), required.")
	_ = c.MarkFlagRequired("file")
	return c
}
