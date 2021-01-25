package providerscan

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func cloudsploitCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "cloudsploit",
		Short: "Run Cloudsploit",
		Long: `Run Cloudsploit against cloud providers.

  $ solube provider-scan aws`,
		Args: cobra.NoArgs,
	}
	c.AddCommand(cloudsploitAWSCommand())
	return c
}

func cloudsploitAWSCommand() *cobra.Command {
	opts := options.PrintOpts{}
	c := &cobra.Command{
		Use:   "aws",
		Short: "Run Cloudsploit for AWS",
		Long: `Scan AWS using Cloudsploit.

No results are sent to Soluble, this is just a convenience utility.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			awsEnvCreds, _ := credentials.NewEnvCredentials().Get()
			homeDir, _ := os.UserHomeDir()
			awsFileCreds, _ := credentials.NewSharedCredentials(filepath.Join(homeDir, ".aws/credentials"), "").Get()
			envvars := make(map[string]string)

			// prefer AWS environment variables to those in ~/.aws/credentials
			switch {
			case awsEnvCreds.HasKeys():
				envvars["AWS_ACCESS_KEY_ID"] = awsEnvCreds.AccessKeyID
				envvars["AWS_SECRET_ACCESS_KEY"] = awsEnvCreds.SecretAccessKey
				envvars["AWS_SESSION_TOKEN"] = awsEnvCreds.SessionToken
			case awsFileCreds.HasKeys():
				envvars["AWS_ACCESS_KEY_ID"] = awsFileCreds.AccessKeyID
				envvars["AWS_SECRET_ACCESS_KEY"] = awsFileCreds.SecretAccessKey
				envvars["AWS_SESSION_TOKEN"] = awsFileCreds.SessionToken
			default:
				return fmt.Errorf("neither envvars nor ~/.aws/credentials file have AWS credentials")
			}
			// and inculde the Soluble API keys
			envvars["SOLUBLE_API_KEY"] = os.Getenv("SOLUBLE_API_KEY")
			envvars["SOLUBLE_ORG_ID"] = os.Getenv("SOLUBLE_ORG_ID")

			if err := hasDocker(); err != nil {
				return fmt.Errorf("cannot run cloudsploit docker container: %w", err)
			}
			// Write the environment variables to a tmpdir with a bindmount
			// (because we don't want to leak sensitive keys into `ps` and logs)
			envFile, err := ioutil.TempFile("", "soluble-cloudsploit")
			if err != nil {
				return fmt.Errorf("unable to create temporary file for cloudsploit: %w", err)
			}
			defer os.Remove(envFile.Name())
			defer envFile.Close()
			for k, v := range envvars {
				if v == "" {
					continue
				}
				_, err := envFile.WriteString(fmt.Sprintf("%s=%s\n", k, v))
				if err != nil {
					return fmt.Errorf("unable to write environment variables to temporary file: %w", err)
				}
			}
			_ = envFile.Sync()

			solubleDir := filepath.Join(homeDir, ".soluble")
			cloudsploitCmd := exec.Command("docker", "run", "-it",
				"--env-file", envFile.Name(),
				"-v", fmt.Sprintf("%s:/app/.soluble:ro", solubleDir), // soluble baseimg handles reading .soluble dir
				"gcr.io/soluble-repo/soluble-cloudsploit:latest",
			) // #nosec G204
			cloudsploitCmd.Stdin = os.Stdin
			cloudsploitCmd.Stdout = os.Stdout
			cloudsploitCmd.Stderr = os.Stderr
			return cloudsploitCmd.Run()
		},
	}
	opts.Register(c)
	return c
}

func hasDocker() error {
	// see: pkg/tools/docker.go
	c := exec.Command("docker", "info")
	err := c.Run()
	switch c.ProcessState.ExitCode() {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("docker server is not available: %w", err)
	case 127:
		return fmt.Errorf("docker is not installed or executable is not in PATH: %w", err)
	default:
		return fmt.Errorf("docker not found or not installed: %w", err)
	}
}
