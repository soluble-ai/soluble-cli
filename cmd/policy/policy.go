package policy

import (
	"os"

	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
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
	var dir string
	c := &cobra.Command{
		Use:   "vet",
		Short: "Vet custom policy for potential errors",
		RunE: func(cmd *cobra.Command, args []string) error {
			m := policy.NewManager(dir)
			err := m.LoadAllRules()
			if err != nil {
				return err
			}
			for ruleType := range m.Rules {
				log.Infof("Found %d {primary:%s} custom rules", len(m.Rules[ruleType]), ruleType)
			}
			return m.ValidateRules()
		},
	}
	c.Flags().StringVarP(&dir, "directory", "d", "", "Validate custom policy in `dir`")
	_ = c.MarkFlagRequired("directory")
	return c
}

func uploadCommand() *cobra.Command {
	var (
		dir     string
		client  options.ClientOpts
		tarball string
		upload  bool
	)
	c := &cobra.Command{
		Use:   "upload",
		Short: "Upload custom policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			if upload {
				if err := client.RequireAPIToken(); err != nil {
					return err
				}
			}
			m := policy.NewManager(dir)
			if err := m.LoadAllRules(); err != nil {
				return err
			}
			if err := m.ValidateRules(); err != nil {
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
			if upload {
				f, err := os.Open(tarball)
				if err != nil {
					return err
				}
				defer f.Close()
				options := []api.Option{
					xcp.WithCIEnv(dir), xcp.WithFileFromReader("tarball", "rules.tar.gz", f),
				}
				_, err = client.GetAPIClient().XCPPost(client.GetOrganization(),
					"policies", nil, nil, options...)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	client.Register(c)
	flags := c.Flags()
	flags.StringVarP(&dir, "directory", "d", "", "Read custom policies from `dir`")
	flags.BoolVar(&upload, "upload", true, "Upload custom policies.  Use --upload=false to disable.")
	flags.StringVar(&tarball, "tarball", "", "Write upload tarball to `file`.  By default the tarball is written to a temporary file.")
	_ = c.MarkFlagRequired("directory")
	return c
}

func testCommand() *cobra.Command {
	var dir string
	c := &cobra.Command{
		Use:   "test",
		Short: "Test custom policy",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, ruleType, rule, target, err := policy.DetectPolicy(dir)
			if err != nil {
				return err
			}
			switch {
			case rule != nil && target != "":
				return m.TestRuleTarget(rule, target)
			case rule != nil:
				return m.TestRule(rule)
			case ruleType != nil:
				return m.TestRuleType(ruleType)
			default:
				return m.TestRules()
			}
		},
	}
	c.Flags().StringVarP(&dir, "directory", "d", "", "Run tests in `dir`")
	return c
}
