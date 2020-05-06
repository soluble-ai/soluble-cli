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

package kubernetes

import (
	"github.com/spf13/cobra"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/options"
)

type EntityOpts struct {
	options.PrintClusterOpts
	kinds       []string
	currentKind string
	allClusters bool
}

func (opts *EntityOpts) Register(c *cobra.Command) {
	opts.ClusterIDOptional = true
	opts.SetFormatter("kind", func(n *jnode.Node, columnName string) string {
		return opts.currentKind
	})
	opts.PrintClusterOpts.Register(c)
	c.Flags().StringSliceVar(&opts.kinds, "kind", []string{}, "The kubernetes entity Kind")
	c.Flags().BoolVar(&opts.allClusters, "all-clusters", false, "Show results from all clusters")
}

func (opts *EntityOpts) GetKinds() []string {
	if len(opts.kinds) == 0 {
		return []string{"Deployment", "Pod", "Node"}
	}
	return opts.kinds
}

func (opts *EntityOpts) GetParams(kind string) map[string]string {
	opts.currentKind = kind
	params := map[string]string{
		"kind": kind,
	}
	clusterID := opts.GetClusterIDParam()
	if clusterID != "" {
		params["clusterId"] = clusterID
	}
	return params
}

func (opts *EntityOpts) GetClusterIDParam() string {
	clusterID := opts.GetClusterID()
	if clusterID != "" && !opts.allClusters {
		return clusterID
	}
	return ""
}

func listKubernetesEntities() *cobra.Command {
	opts := EntityOpts{
		PrintClusterOpts: options.PrintClusterOpts{
			PrintOpts: options.PrintOpts{
				Path:    []string{"data"},
				Columns: []string{"kind", "name", "namespace", "creationTimestamp", "updateTs+", "cluster"},
			},
		},
	}
	opts.AddTransformer(func(n *jnode.Node) *jnode.Node {
		c := n.Path("clusterDisplayName").AsText()
		if c == "" {
			c = n.Path("clusterId").AsText()
		}
		n.Put("cluster", c)
		return n
	})
	c := &cobra.Command{
		Use:   "list",
		Short: "List kubernetes entities",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient := opts.GetAPIClient()
			opts.PrintHeader()
			for _, kind := range opts.GetKinds() {
				result, err := apiClient.GetWithParams("org/{org}/kubernetes/entities",
					opts.GetParams(kind))
				if err != nil {
					return err
				}
				opts.PrintRows(result)
			}
			opts.Flush()

			return nil
		},
	}
	opts.Register(c)
	return c
}

func getEntityCommand() *cobra.Command {
	var opts options.PrintClusterOpts
	var kind string
	var name string
	var namespace string
	c := &cobra.Command{
		Use:   "get namespace/name",
		Short: "Get a kubernetes entity",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient := opts.GetAPIClient()
			params := map[string]string{
				"kind":      kind,
				"name":      name,
				"namespace": namespace,
			}
			result, err := apiClient.GetWithParams("org/{org}/kubernetes/entity", params)
			if err != nil {
				return err
			}
			for _, n := range result.Path("data").Elements() {
				data := n.Path("data")
				if !data.IsMissing() {
					obj, _ := jnode.FromJSON([]byte(data.AsText()))
					if obj.IsObject() {
						opts.PrintResultYAML(obj)
						continue
					}
				}
				opts.PrintResultYAML(data)
			}
			return nil
		},
	}
	opts.ClusterIDOptional = true
	opts.Register(c)
	c.Flags().StringVar(&kind, "kind", "", "The kubernetes entity Kind")
	c.Flags().StringVar(&name, "name", "", "The name of the object")
	c.Flags().StringVar(&namespace, "namespace", "default", "The namespace of the object")
	_ = c.MarkFlagRequired("kind")
	_ = c.MarkFlagRequired("name")
	return c
}

func Command() *cobra.Command {
	k := &cobra.Command{
		Use:     "kubernetes",
		Aliases: []string{"k8s"},
		Short:   "Commands for kubernetes clusters",
	}
	k.AddCommand(listKubernetesEntities())
	k.AddCommand(getEntityCommand())
	return k
}
