package opal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/archive"
	"github.com/soluble-ai/soluble-cli/pkg/exit"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/policy"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	tfutil "github.com/soluble-ai/soluble-cli/pkg/tools/util"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
	InputType            string
	varFiles             []string
	ExtraArgs            []string
	EnableModuleDownload bool
	iacPlatform          tools.IACPlatform
}

var _ tools.Single = (*Tool)(nil)

func (t *Tool) Name() string {
	return "opal"
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:    "opal",
		Short:  "Scan IAC for security issues",
		Hidden: true,
	}
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringSliceVar(&t.varFiles, "var-file", nil, "Pass additional variable `files` to opal")
	flags.BoolVar(&t.EnableModuleDownload, "enable-module-download", false, "Use --enable-module-download=true to enable.")
}

func (t *Tool) setVarFiles() error {
	for i, varFile := range t.varFiles {
		varFile, err := filepath.Abs(varFile)
		if err != nil {
			return err
		}
		if !util.FileExists(varFile) {
			return fmt.Errorf("var file %s does not exist", varFile)
		}
		t.varFiles[i] = varFile
	}
	return nil
}

func (t *Tool) GetVarFiles() []string {
	return t.varFiles
}

func (t *Tool) Validate() error {
	switch t.InputType {
	case "arm":
		t.iacPlatform = tools.ARM
	case "k8s":
		t.iacPlatform = tools.Kubernetes
	case "cfn":
		t.iacPlatform = tools.Cloudformation
	case "tf-plan":
		t.iacPlatform = tools.TerraformPlan
	case "tf":
		fallthrough
	case "":
		t.iacPlatform = tools.Terraform
	default:
		return fmt.Errorf("opal does not support %s", t.InputType)
	}
	err := t.setVarFiles()
	if err != nil {
		return err
	}
	return t.DirectoryBasedToolOpts.Validate()
}

func (t *Tool) Run() (*tools.Result, error) {
	result := &tools.Result{
		Directory:   t.GetDirectory(),
		IACPlatform: t.iacPlatform,
	}

	d, err := t.InstallTool(&download.Spec{
		Name: "opal",
		URL:  "github.com/lacework/opal-releases",
	})
	if err != nil {
		log.Errorf("Could not run opal scan %s", err)
		os.Exit(0)
	}

	// create a unique temp dir for downloaded policies
	customPoliciesDir, err := util.GetTempDirPath()
	if err != nil {
		return nil, err
	}

	if t.EnableModuleDownload {
		err := tfutil.RunTerraformGet(t.GetDirectory(), t.RunOpts)
		if err != nil {
			log.Warnf("{warning:terraform get} failed ")
			result.AddValue("TERRAFORM_GET_FAILED", "true")
		}
	}

	// if CustomPoliciesDir is present prepare those policies and use them for a local assessment
	// overriding upload flag and PreparedCustomPoliciesDir
	if t.CustomPoliciesDir != "" {
		store := policy.NewStore(t.CustomPoliciesDir, false)
		preparedPoliciesDir, err := os.MkdirTemp("", "policy*")
		if err != nil {
			return nil, err
		}
		t.PreparedCustomPoliciesDir = preparedPoliciesDir
		exit.AddFunc(func() { _ = os.RemoveAll(preparedPoliciesDir) })
		if err := store.LoadPoliciesOfType(policy.GetPolicyType("opal")); err != nil {
			return nil, err
		}
		if err := store.PreparePolicies(preparedPoliciesDir); err != nil {
			return nil, err
		}
		t.AssessmentOpts.UploadEnabled = false
	}
	// if the PreparedCustomPoliciesDir is set, policies are loaded from here and not downloaded
	preparedPoliciesDir := t.PreparedCustomPoliciesDir
	if preparedPoliciesDir == "" {
		apiClient, err := t.GetAPIClient()
		if err != nil {
			return nil, err
		}
		err = DownloadPolicies(apiClient, customPoliciesDir, t.AssessmentOpts)
		if err != nil {
			if errors.Is(err, api.ErrNoContent) {
				log.Infof("There are no custom policies available for {primary:%s} ", t.Name())
				return &tools.Result{}, nil
			}
			return nil, err
		}
		store := policy.NewDownloadStore(customPoliciesDir)
		preparedPoliciesDir, err = os.MkdirTemp("", "policy*")
		if err != nil {
			return nil, err
		}
		exit.AddFunc(func() { _ = os.RemoveAll(preparedPoliciesDir) })
		if err := store.LoadPoliciesOfType(policy.GetPolicyType("opal")); err != nil {
			return nil, err
		}
		// policies from the customPoliciesDir download dir are prepared in the preparedPoliciesDir
		if err := store.PreparePolicies(preparedPoliciesDir); err != nil {
			return nil, err
		}
		md, err := store.GetPolicyUploadMetadata("policies-upload-metadata.json")
		if err != nil {
			return nil, err
		}
		// used in the final assessment
		t.AssessmentOpts.CustomPolicyMetadata = md
		lmd, err := store.GetPolicyUploadMetadata("lacework-policies-upload-metadata.json")
		if err != nil {
			return nil, err
		}
		t.AssessmentOpts.LaceworkPolicyMetadata = lmd
	}

	args := []string{"run", "--format", "json"}
	// include downloaded custom policies or local prepared policies in the scan
	if preparedPoliciesDir != "" {
		args = append(args, "--include", preparedPoliciesDir)
	}
	if t.InputType != "" {
		args = append(args, "--input-type", t.InputType)
	}
	for _, varFile := range t.GetVarFiles() {
		args = append(args, "--var-file", varFile)
	}
	args = append(args, t.ExtraArgs...)
	args = append(args, ".")
	// #nosec G204
	c := exec.Command(d.GetExePath("opal"), args...)
	c.Dir = t.GetDirectory()
	c.Stderr = os.Stderr
	executeCommand := t.ExecuteCommand(c)
	result.ExecuteResult = executeCommand
	if !executeCommand.ExpectExitCode(0, 1) {
		return result, nil
	}
	n, ok := executeCommand.ParseJSON()
	if !ok {
		return result, nil
	}
	t.parseResults(result, n)
	return result, nil
}

func (t *Tool) parseResults(result *tools.Result, n *jnode.Node) {
	result.Data = n
	for _, rr := range n.Path("policy_results").Elements() {
		loc := rr.Path("source_location").Get(0)
		result.Findings = append(result.Findings, &assessments.Finding{
			Severity: rr.Path("policy_severity").AsText(),
			Pass:     rr.Path("policy_result").AsText() == "PASS",
			FilePath: loc.Path("path").AsText(),
			Line:     loc.Path("line").AsInt(),
			Title:    rr.Path("policy_summary").AsText(),
			Tool: map[string]string{
				"policy_id": rr.Path("policy_id").AsText(),
			},
			SID: rr.Path("policy_id").AsText(),
		})
	}
}

func DownloadPolicies(apiClient *api.Client, dir string, o tools.AssessmentOpts) error {
	err := os.MkdirAll(dir, 0744)
	if err != nil {
		return err
	}
	log.Debugf(fmt.Sprintf("Policies downloaded to directory: %s", dir))
	policyZipPath := filepath.Join(dir, "policies.zip")
	data, err := apiClient.Download(fmt.Sprintf("/api/v1/org/%s/policies/opal/policies.zip", apiClient.Organization))
	if err != nil {
		return err
	}
	body := bytes.NewReader(data)
	policyZipFile, err := os.Create(policyZipPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(policyZipFile, body)
	if err != nil {
		return err
	}
	// check there is a zip
	err = archive.Do(archive.Unzip, policyZipPath, dir, nil)
	if err != nil {
		return err
	}
	if o.DisableCustomPolicies && o.PreparedCustomPoliciesDir == "" {
		// assume disable custom policies does not mean disable lacework policies
		err = tools.ExtractArchives(dir, []string{"lacework_policies.zip"})
	} else {
		err = tools.ExtractArchives(dir, []string{"policies.zip", "lacework_policies.zip"})
	}

	if err != nil {
		return err
	}
	return err
}
