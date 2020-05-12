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

	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/spf13/cobra"
)

type ClusterOpts struct {
	ClientOpts
	ClusterID         string
	ClusterIDOptional bool
	SetDefault        bool
}

var _ Interface = &ClusterOpts{}

func (opts *ClusterOpts) SetContextValues(context map[string]string) {
	opts.ClientOpts.SetContextValues(context)
	context["clusterID"] = opts.GetClusterID()
}

func (opts *ClusterOpts) Register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&opts.ClusterID, "cluster-id", "", "The cluster id.")
	cmd.Flags().BoolVar(&opts.SetDefault, "set-default", false, "Set the default cluster-id to the value --cluster-id")
	opts.ClientOpts.Register(cmd)
	AddPreRunE(cmd, func(cmd *cobra.Command, args []string) error {
		flag := cmd.Flag("cluster-id")
		if flag.Value.String() == "" && !opts.ClusterIDOptional {
			clusterID := config.Config.DefaultClusterID
			if clusterID == "" {
				return fmt.Errorf("no --cluster-id flag given, and no default cluster set (set a default cluster with 'config set defaultclusterid')")
			}
			_ = flag.Value.Set(clusterID)
		}
		if opts.SetDefault && flag.Value.String() != "" {
			config.Config.DefaultClusterID = flag.Value.String()
			config.Save()
		}
		return nil
	})
}

func (opts *ClusterOpts) GetClusterID() string {
	return opts.ClusterID
}
