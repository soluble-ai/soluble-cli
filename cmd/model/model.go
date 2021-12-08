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

package model

import (
	"fmt"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/model"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "model",
		Short: "Manage API models",
	}
	c.AddCommand(listModels())
	c.AddCommand(addModel())
	return c
}

func listModels() *cobra.Command {
	opts := options.PrintOpts{
		Path:    []string{"models"},
		Columns: []string{"command", "location", "version"},
	}
	c := &cobra.Command{
		Use:   "list",
		Short: "List API models",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			result := jnode.NewObjectNode()
			a := result.PutArray("models")
			for _, model := range model.Models {
				a.AppendObject().Put("location", model.FileName).
					Put("command", model.Command.Name).
					Put("version", model.Version)
			}
			opts.PrintResult(result)
		},
	}
	opts.Register(c)
	return c
}

func addModel() *cobra.Command {
	opts := options.PrintOpts{}
	var url string
	c := &cobra.Command{
		Use:   "add-git-location",
		Short: "Add an API model from git",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !strings.HasPrefix(url, "git@") {
				return fmt.Errorf("only git@... urls are supported")
			}
			if util.StringSliceContains(config.GlobalConfig.ModelLocations, url) {
				return nil
			}
			source, err := model.GetGitSource(url)
			if err != nil {
				return err
			}
			log.Infof("Added model source {info:%s}", source)
			config.GlobalConfig.ModelLocations = append(config.GlobalConfig.ModelLocations, url)
			return config.Save()
		},
	}
	opts.Register(c)
	c.Flags().StringVar(&url, "url", "", "Add models in this git repository")
	_ = c.MarkFlagRequired("url")
	return c
}
