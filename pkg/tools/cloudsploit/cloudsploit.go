// Copyright 2021 Soluble Inc
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

package cloudsploit

import (
	"context"
	"fmt"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	opts := &tools.RunOpts{}
	c := &cobra.Command{
		Use:   "cloudsploit",
		Short: "Scan cloud infrastructure with Cloudsploit",
		RunE: func(cmd *cobra.Command, args []string) error {
			awscfg, err := awsconfig.LoadDefaultConfig(context.Background())
			if err != nil {
				return err
			}
			creds, err := awscfg.Credentials.Retrieve(context.Background())
			if err != nil {
				return err
			}
			api, err := opts.GetAPIClient()
			if err != nil {
				return err
			}
			env := map[string]string{}
			env["AWS_ACCESS_KEY_ID"] = creds.AccessKeyID
			env["AWS_SECRET_ACCESS_KEY"] = creds.SecretAccessKey
			env["AWS_SESSION_TOKEN"] = creds.SessionToken
			env["SOLUBLE_API_SERVER"] = api.APIServer
			env["SOLUBLE_API_TOKEN"] = api.LegacyAPIToken
			envFile, err := writeEnvFile(env)
			if err != nil {
				return err
			}
			defer func() { _ = os.Remove(envFile) }()
			docker := &tools.DockerTool{
				Name:  "cloudsploit",
				Image: "gcr.io/soluble-repo/soluble-cloudsploit:latest",
				DockerArgs: []string{"--env-file", envFile,
					"-v", fmt.Sprintf("%s:/app/.solulble:ro", config.ConfigDir)},
				Args:   args,
				Stdout: os.Stdout,
			}
			_, err = opts.RunDocker(docker)
			return err
		},
	}
	opts.Register(c)
	return c
}

func writeEnvFile(env map[string]string) (string, error) {
	// Write the environment variables to a tmpdir with a bindmount
	// (because we don't want to leak sensitive keys into `ps` and logs)
	envFile, err := os.CreateTemp("", "soluble-cloudsploit*")
	if err != nil {
		return "", fmt.Errorf("unable to create temporary file for cloudsploit: %w", err)
	}
	defer envFile.Close()
	for k, v := range env {
		if v == "" {
			continue
		}
		fmt.Fprintf(envFile, "%s=%s\n", k, v)
	}
	return envFile.Name(), nil
}
