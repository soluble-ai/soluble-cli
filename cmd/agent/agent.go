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

package agent

import (
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/go-resty/resty/v2"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/model"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func pingClusterCmd() *cobra.Command {
	var opts options.PrintClusterOpts
	c := &cobra.Command{
		Use:   "ping",
		Short: "Send a ping message the cluster operator",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterID := opts.GetClusterID()
			path := fmt.Sprintf("org/{org}/clusters/%s/ping", clusterID)
			apiClient := opts.GetAPIClient()
			result, err := apiClient.Post(path, jnode.NewObjectNode())
			if err != nil {
				return err
			}
			opts.PrintResult(result)

			messageID := result.Path("messageId").AsText()

			if messageID == "" {
				return fmt.Errorf("ping returned empty messageId")
			}

			path = fmt.Sprintf("org/{org}/clusters/%s/ping/%s", clusterID, messageID)
			return retry.Do(
				func() error {
					result, err := apiClient.Get(path)
					if err != nil {
						log.Warnf("%s", err)
						return err
					}
					opts.PrintResult(result)
					return nil
				}, retry.Delay(10*time.Second), retry.DelayType(retry.FixedDelay),
				retry.LastErrorOnly(true), retry.Attempts(6),
			)
		},
	}
	opts.Register(c)
	return c
}

func deployAgentCmd() *cobra.Command {
	opts := options.ClientOpts{}
	var directory string
	var kustomize bool
	c := &cobra.Command{
		Use:   "deploy",
		Short: "Download the kubenertes resources to deploy the agent in a cluster",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient := opts.GetAPIClient()
			var tail string

			if kustomize {
				tail = "soluble-agent.zip"
			} else {
				tail = "soluble-agent.yml"
			}
			path := fmt.Sprintf("org/{org}/config/%s", tail)
			if directory != "" {
				apiClient.GetClient().SetOutputDirectory(directory)
			}
			_, err := apiClient.Get(path, func(req *resty.Request) {
				req.SetOutput(tail)
			})
			if err != nil {
				return err
			}
			var file string
			if directory == "" {
				file = tail
			} else {
				file = fmt.Sprintf("%s/%s", directory, tail)
			}
			log.Infof("Wrote result to {primary:%s}", file)
			if !kustomize {
				log.Infof("Run {primary:kubectl create --save-config -f %s} to apply", file)
			} else {
				var cmd string
				if directory != "" {
					cmd += fmt.Sprintf("cd %s && ", directory)
				}
				cmd += fmt.Sprintf("unzip %s && kubectl apply -k .", tail)
				log.Infof("Run {info:%s} to apply", cmd)
			}
			return nil
		},
	}
	opts.Register(c)
	c.Flags().StringVarP(&directory, "directory", "d", "", "Output directory.")
	c.Flags().BoolVar(&kustomize, "kustomize", false, "Download a zip file with a kustomization.yaml manifest")
	return c
}

func Command() *cobra.Command {
	agent := &cobra.Command{
		Use:   "agent",
		Short: "Manage agents",
	}
	agent.AddCommand(pingClusterCmd())
	agent.AddCommand(deployAgentCmd())
	return agent
}

func firstMessageColumnFunction(n *jnode.Node) interface{} {
	// pick the first message, and if it's JSON use the "msg" field
	s := n.Path("messages").Get(0).Path("message").AsText()
	if m, err := jnode.FromJSON([]byte(s)); err == nil {
		return m.Path("msg")
	}
	return ""
}

func init() {
	model.RegisterColumnFunction("first_message", firstMessageColumnFunction)
}
