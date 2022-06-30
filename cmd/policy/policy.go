package policy

import (
	"os"

	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "policy",
		Short: "Custom policy management",
	}
	c.AddCommand(
		vetCommand(),
		uploadCommand(),
		testCommand(),
	)
	return c
}

func vetCommand() *cobra.Command {
	m := &policy.Manager{}
	c := &cobra.Command{
		Use:   "vet",
		Short: "Vet custom policy for potential errors",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := m.LoadRules()
			if err != nil {
				return err
			}
			for ruleType := range m.Rules {
				log.Infof("Found %d {info:%s} custom rules", len(m.Rules[ruleType]), ruleType)
			}
			metrics, err := m.ValidateRules()
			log.Infof("Validated {primary:%d} custom rules", metrics.Count)
			return err
		},
	}
	m.Register(c)
	return c
}

func uploadCommand() *cobra.Command {
	var (
		m          policy.Manager
		tarball    string
		uploadOpts tools.UploadOpts
	)
	c := &cobra.Command{
		Use:   "upload",
		Short: "Upload custom policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			if uploadOpts.UploadEnabled {
				if err := m.RequireAPIToken(); err != nil {
					return err
				}
			}
			if err := m.LoadRules(); err != nil {
				return err
			}
			if _, err := m.ValidateRules(); err != nil {
				return err
			}
			if tarball == "" {
				var err error
				tarball, err = util.TempFile("rules*.tar.gz")
				if err != nil {
					return err
				}
				defer os.Remove(tarball)
			}
			if err := m.CreateTarBall(tarball); err != nil {
				return err
			}
			if uploadOpts.UploadEnabled {
				f, err := os.Open(tarball)
				if err != nil {
					return err
				}
				defer f.Close()
				options := []api.Option{
					xcp.WithCIEnv(m.Dir),
					xcp.WithFileFromReader("tarball", "rules.tar.gz", f),
				}
				options = uploadOpts.AppendUploadOptions(m.Dir, options)
				_, err = m.GetAPIClient().XCPPost(m.GetOrganization(),
					"custom/policy", nil, nil, options...)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	m.Register(c)
	flags := c.Flags()
	uploadOpts.DefaultUploadEnabled = true
	uploadOpts.Register(c)
	flags.StringVar(&tarball, "save-tarball", "", "Save the upload tarball to `file`.  By default the tarball is written to a temporary file.")
	flags.Lookup("upload").Usage = "Upload rules to lacework.  Use --upload=false to skip uploading."
	flags.Lookup("upload-errors").Hidden = true // doesn't make sense here
	_ = c.MarkFlagRequired("directory")
	return c
}

func testCommand() *cobra.Command {
	m := &policy.Manager{}
	c := &cobra.Command{
		Use:   "test",
		Short: "Test custom policy",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := m.DetectPolicy()
			if err != nil {
				return err
			}
			metrics, err := m.TestRules()
			if metrics.FailureCount == 0 {
				log.Infof("Ran {primary:%d} tests and all passed", metrics.TestCount)
			} else {
				log.Infof("Ran {primary:%d} tests with {success:%d} passed and {danger:%d} failed",
					metrics.TestCount, metrics.TestCount-metrics.FailureCount, metrics.FailureCount)
			}
			return err
		},
	}
	m.Register(c)
	return c
}
