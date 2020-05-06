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

package cluster

import (
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "cluster",
		Short: "Commands to control managed clusters",
	}
	c.AddCommand(newListClustersCommand())
	return c
}

func newListClustersCommand() *cobra.Command {
	opts := options.PrintClusterOpts{
		PrintOpts: options.PrintOpts{
			Path: []string{"data"},
			Columns: []string{
				"default", "displayName", "clusterId", "clusterEndpoint", "updateTs+", "clusterManager", "kubeGitVersion", "agentVersion",
			},
			Formatters: map[string]options.Formatter{
				"default": func(n *jnode.Node, columnName string) string {
					if isDefaultClusterID(n.Path("clusterId").AsText()) {
						return "   *"
					}
					return ""
				},
			},
			SortBy: []string{"displayName"},
		},
		ClusterOpts: options.ClusterOpts{
			// TODO - not optional, actually not accepted
			ClusterIDOptional: true,
		},
	}
	c := &cobra.Command{
		Use:   "list",
		Short: "List clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := opts.GetAPIClient().Get("org/{org}/clusters")
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

func isDefaultClusterID(clusterID string) bool {
	return clusterID == config.Config.DefaultClusterID
}
