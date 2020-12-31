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
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

// OptionsCommand is a command based on the options framework
type OptionsCommand struct {
	options.Interface
	ClientOpts   *options.ClientOpts
	PrintOpts    *options.PrintOpts
	ClusterOpts  *options.ClusterOpts
	CobraCommand *cobra.Command
}

func (w *OptionsCommand) Initialize(c *cobra.Command, cm *CommandModel) Command {
	w.CobraCommand = c
	if w.ClientOpts != nil {
		opts := w.ClientOpts
		opts.AuthNotRequired = cm.AuthNotRequired != nil && *cm.AuthNotRequired
		opts.APIPrefix = cm.model.APIPrefix
		if cm.DefaultTimeout != nil {
			opts.DefaultTimeout = *cm.DefaultTimeout
		}
	}
	if w.PrintOpts != nil {
		opts := w.PrintOpts
		if cm.Result != nil {
			r := *cm.Result
			if r.Path != nil {
				opts.Path = *r.Path
				opts.Columns = *r.Columns
				if r.DiffColumn != nil {
					opts.DiffColumn = *r.DiffColumn
				}
				if r.VersionColumn != nil {
					opts.VersionColumn = *r.VersionColumn
				}
			}
			if r.WideColumns != nil {
				opts.WideColumns = *r.WideColumns
			}
			if r.Sort != nil {
				opts.DefaultSortBy = *r.Sort
			}
			if r.Formatters != nil {
				for columnName := range *r.Formatters {
					formatter := r.GetFormatter(columnName)
					opts.SetFormatter(columnName, formatter)
				}
			}
			if r.ComputedColumns != nil {
				for columnName := range *r.ComputedColumns {
					computer := r.GetColumnFunction(columnName)
					opts.SetColumnFunction(columnName, computer)
				}
			}
			if r.DefaultOutputFormat != nil {
				opts.DefaultOutputFormat = *r.DefaultOutputFormat
			}
		}
	}
	if w.ClusterOpts != nil {
		opts := w.ClusterOpts
		opts.ClusterIDOptional = cm.ClusterIDOptional != nil && *cm.ClusterIDOptional
	}
	w.Interface.Register(c)
	return w
}

func (w *OptionsCommand) GetAPIClient() *api.Client {
	if w.ClientOpts != nil {
		return w.ClientOpts.GetAPIClient()
	}
	return nil
}

func (w *OptionsCommand) GetUnauthenticatedAPIClient() *api.Client {
	if w.ClientOpts != nil {
		return w.ClientOpts.GetUnauthenticatedAPIClient()
	}
	return nil
}

func (w *OptionsCommand) PrintResult(n *jnode.Node) {
	if w.PrintOpts != nil {
		w.PrintOpts.PrintResult(n)
	}
}

func (w *OptionsCommand) GetCobraCommand() *cobra.Command {
	return w.CobraCommand
}

func init() {
	RegisterCommandType("print_client", func(c *cobra.Command, cm *CommandModel) Command {
		opts := &options.PrintClientOpts{}
		w := &OptionsCommand{
			Interface:  opts,
			ClientOpts: &opts.ClientOpts,
			PrintOpts:  &opts.PrintOpts,
		}
		return w.Initialize(c, cm)
	})
	RegisterCommandType("print_cluster", func(c *cobra.Command, cm *CommandModel) Command {
		opts := &options.PrintClusterOpts{}
		w := &OptionsCommand{
			Interface:   opts,
			ClientOpts:  &opts.ClientOpts,
			PrintOpts:   &opts.PrintOpts,
			ClusterOpts: &opts.ClusterOpts,
		}
		return w.Initialize(c, cm)
	})
}
