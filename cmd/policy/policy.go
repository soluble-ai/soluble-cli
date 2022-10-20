package policy

import (
	"os"

	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/policy/custompolicybuilder"
	"github.com/soluble-ai/soluble-cli/pkg/policy/manager"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
	"github.com/spf13/cobra"

	_ "github.com/soluble-ai/soluble-cli/pkg/policy/checkov"
	_ "github.com/soluble-ai/soluble-cli/pkg/policy/opal"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:    "policy",
		Short:  "Custom policy management",
		Hidden: true,
	}
	c.AddCommand(
		vetCommand(),
		uploadCommand(),
		testCommand(),
		CreateCommand(),
	)
	return c
}

func CreateCommand() *cobra.Command {
	// only available for opal
	cpb := &custompolicybuilder.PolicyTemplate{Tool: "opal"}
	c := &cobra.Command{
		Use:   "create",
		Short: "Create custom policy. Generates skeleton policy and metadata file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cpb.PromptInput(); err != nil {
				return err
			}
			if err := cpb.CreateCustomPolicyTemplate(); err != nil {
				return err
			}
			return nil
		},
	}
	return c
}

func vetCommand() *cobra.Command {
	m := &manager.M{}
	c := &cobra.Command{
		Use:   "vet",
		Short: "Vet custom policy for potential errors",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := m.DetectPolicy(""); err != nil {
				return err
			}
			result := m.ValidatePolicies()
			if result.Errors != nil {
				return result.Errors
			}
			log.Infof("Validated {primary:%d} custom policies", result.Valid+result.Invalid)
			m.MustPrintStructResult(result)
			return nil
		},
	}
	m.Register(c)
	return c
}

func uploadCommand() *cobra.Command {
	var (
		m          manager.M
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
			if err := m.LoadPolicies(); err != nil {
				return err
			}
			if res := m.ValidatePolicies(); res.Errors != nil {
				return res.Errors
			}
			if tarball == "" {
				var err error
				tarball, err = util.TempFile("policies*.tar.gz")
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
					xcp.WithFileFromReader("tarball", "policies.tar.gz", f),
				}
				options = uploadOpts.AppendUploadOptions(m.Dir, options)
				api, err := m.GetAPIClient()
				if err != nil {
					return err
				}
				_, err = api.XCPPost("custom/policy", nil, nil, options...)
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
	flags.Lookup("upload").Usage = "Upload policies to lacework.  Use --upload=false to skip uploading."
	flags.Lookup("upload-errors").Hidden = true // doesn't make sense here
	_ = c.MarkFlagRequired("directory")
	return c
}

func testCommand() *cobra.Command {
	m := &manager.M{}
	c := &cobra.Command{
		Use:   "test",
		Short: "Test custom policy",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := m.DetectPolicy(""); err != nil {
				return err
			}
			if res := m.ValidatePolicies(); res.Errors != nil {
				return res.Errors
			}
			metrics, err := m.TestPolicies()
			if metrics.Failed == 0 {
				log.Infof("Ran {primary:%d} tests and all passed", metrics.Passed)
			} else {
				log.Infof("Ran {primary:%d} tests with {success:%d} passed and {danger:%d} failed",
					metrics.Passed+metrics.Failed, metrics.Passed, metrics.Failed)
			}
			m.MustPrintStructResult(metrics)
			return err
		},
	}
	m.Register(c)
	return c
}
