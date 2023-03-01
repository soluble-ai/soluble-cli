package policy

import (
	"fmt"
	"os"
	"time"

	"github.com/soluble-ai/soluble-cli/pkg/policy/policyimporter"

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
		createCommand(),
		opalConvertCommand(),
		prepareCommand(),
		downloadCommand(),
	)
	return c
}

func createCommand() *cobra.Command {
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

func opalConvertCommand() *cobra.Command {
	// only available for opal
	c := &cobra.Command{
		Use:   "convert",
		Short: "Restructure and generate metadata for opal built-in policies to fit lacework directory structure.",
		// opalConvertCommand is for internal use only
		// for conversion of regula policies to lacework policies and is not supported.
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			converter := &policyimporter.Converter{}
			if err := converter.PromptInput(); err != nil {
				return err
			}
			if err := converter.ConvertOpalBuiltIns(); err != nil {
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

func prepareCommand() *cobra.Command {
	m := &manager.M{}
	var dir string
	c := &cobra.Command{
		Use:     "prepare",
		Aliases: []string{"prop"},
		Short:   "Generated the prepared policy for evaluation by the underlying engine",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := m.DetectPolicy(""); err != nil {
				return err
			}
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			return m.PreparePolicies(dir)
		},
	}
	m.Register(c)
	c.Flags().StringVar(&dir, "output-dir", "prepared", "The directory to generate the policies into.")
	return c
}

func uploadCommand() *cobra.Command {
	var (
		m          manager.M
		zipArchive string
		uploadOpts tools.UploadOpts
		allowEmpty bool
	)
	c := &cobra.Command{
		Use:   "upload",
		Short: "Upload custom policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			if uploadOpts.UploadEnabled {
				if err := m.RequireAuthentication(); err != nil {
					return err
				}
			}
			if err := m.LoadPolicies(); err != nil {
				return err
			}
			if len(m.Policies) == 0 && !allowEmpty {
				return fmt.Errorf("no policies found." +
					"\n\t - Ensure path provided points to the parent directory of the /policies directory" +
					"\n\t - or use --allow-empty to upload no policies.")
			}
			if res := m.ValidatePolicies(); res.Errors != nil {
				return res.Errors
			}
			if zipArchive == "" {
				var err error
				zipArchive, err = util.TempFile("policies*.zip")
				if err != nil {
					return err
				}
				defer os.Remove(zipArchive)
			}
			if err := m.CreateZipArchive(zipArchive); err != nil {
				return err
			}
			if uploadOpts.UploadEnabled {
				f, err := os.Open(zipArchive)
				if err != nil {
					return err
				}
				defer f.Close()
				options := []api.Option{
					xcp.WithCIEnv(m.Dir),
					xcp.WithFileFromReader("archive", "policies.zip", f),
				}
				options = uploadOpts.AppendUploadOptions(m.Dir, options)
				apiClient, err := m.GetAPIClient()
				if err != nil {
					return err
				}
				_, err = apiClient.XCPPost("custom/policy", nil, nil, options...)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	m.RegisterUpload(c)
	flags := c.Flags()
	uploadOpts.DefaultUploadEnabled = true
	uploadOpts.Register(c)
	flags.StringVar(&zipArchive, "save-zip-file", "", "Save the upload zip archive to `file`.  By default the archive is written to a temporary file.")
	flags.BoolVar(&allowEmpty, "allow-empty", false, "Allow upload of no policies.")
	flags.Lookup("upload").Usage = "Upload policies to lacework.  Use --upload=false to skip uploading."
	flags.Lookup("upload-errors").Hidden = true // doesn't make sense here
	return c
}

func downloadCommand() *cobra.Command {
	var (
		m manager.M
	)
	c := &cobra.Command{
		Use:   "download",
		Short: "Download Lacework and custom opal policies.",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := m.GetAPIClient()
			if err != nil {
				return err
			}
			if apiClient.LegacyAPIToken == "" && apiClient.LaceworkAPIToken == "" {
				return nil
			}
			url := "/api/v1/org/{org}/policies/opal/policies.zip"
			d, err := m.InstallAPIServerArtifact(fmt.Sprintf("opal-%s-policies",
				apiClient.Organization), url, 1*time.Minute)
			if err != nil {
				return err
			}
			err = tools.ExtractArchives(d.Dir, []string{"policies.zip", "lacework_policies.zip"})
			if err != nil {
				return err
			}
			return nil
		},
	}
	m.RegisterDownload(c)
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
