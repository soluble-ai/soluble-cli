package providerscan

import (
	"fmt"
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
			envCreds, _ := credentials.NewEnvCredentials().Get()
			homeDir, _ := os.UserHomeDir()
			fileCreds, _ := credentials.NewSharedCredentials(filepath.Join(homeDir, ".aws/credentials"), "default").Get()
			// prefer envvars
			var (
				awsAccessKeyID     string
				awsSecretAccessKey string
				awsSessionToken    string
			)
			switch {
			case envCreds.HasKeys():
				awsAccessKeyID = envCreds.AccessKeyID
				awsSecretAccessKey = envCreds.SecretAccessKey
				awsSessionToken = envCreds.SessionToken
			case fileCreds.HasKeys():
				awsAccessKeyID = fileCreds.AccessKeyID
				awsSecretAccessKey = fileCreds.SecretAccessKey
				awsSessionToken = fileCreds.SessionToken
			default:
				return fmt.Errorf("neither envvars nor ~/.aws/credentials file have AWS credentials")
			}
			if err := hasDocker(); err != nil {
				return fmt.Errorf("cannot run cloudsploit: %w", err)
			}
			solubleDir := filepath.Join(homeDir, ".soluble")
			cloudsploitCmd := exec.Command("docker", "run", "-it",
				"-e", fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", awsAccessKeyID),
				"-e", fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", awsSecretAccessKey),
				"-e", fmt.Sprintf("AWS_SESSION_TOKEN=%s", awsSessionToken),
				"-e", "SOLUBLE_API_KEY", // can come from ENV...
				"-e", "SOLUBLE_ORG_ID",
				"-v", fmt.Sprintf("%s:/app/.soluble", solubleDir), // ...or from file
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
