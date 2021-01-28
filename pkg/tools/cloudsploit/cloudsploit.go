package cloudsploit

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws/credentials"
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
			creds, err := credentials.NewChainCredentials([]credentials.Provider{
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
			}).Get()
			if err != nil {
				return err
			}
			env := map[string]string{}
			env["AWS_ACCESS_KEY_ID"] = creds.AccessKeyID
			env["AWS_SECRET_ACCESS_KEY"] = creds.SecretAccessKey
			env["AWS_SESSION_TOKEN"] = creds.SessionToken
			env["SOLUBLE_API_SERVER"] = opts.GetAPIClientConfig().APIServer
			env["SOLUBLE_API_TOKEN"] = opts.GetAPIClientConfig().APIToken
			envFile, err := writeEnvFile(env)
			if err != nil {
				return err
			}
			defer func() { _ = os.Remove(envFile) }()
			docker := &tools.DockerTool{
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
	envFile, err := ioutil.TempFile("", "soluble-cloudsploit*")
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
