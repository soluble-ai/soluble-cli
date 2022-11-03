package opal

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
	InputType string
	VarFiles  []string
	ExtraArgs []string

	iacPlatform tools.IACPlatform
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
	flags.StringSliceVar(&t.VarFiles, "var-file", nil, "Pass additional variable `files` to opal")
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
	for _, varFile := range t.VarFiles {
		if !util.FileExists(varFile) {
			return fmt.Errorf("var file %s does not exist", varFile)
		}
	}
	return t.DirectoryBasedToolOpts.Validate()
}

func (t *Tool) Run() (*tools.Result, error) {
	result := &tools.Result{
		Directory:   t.GetDirectory(),
		IACPlatform: t.iacPlatform,
	}
	d, err := t.InstallTool(&download.Spec{Name: "opal"})
	if err != nil {
		return nil, err
	}
	customPoliciesDir, err := t.GetCustomPoliciesDir("opal")
	if err != nil {
		return nil, err
	}
	args := []string{"run", "--format", "json"}
	if customPoliciesDir != "" {
		args = append(args, "--include", customPoliciesDir)
	}
	if t.InputType != "" {
		args = append(args, "--input-type", t.InputType)
	}
	for _, varFile := range t.VarFiles {
		args = append(args, "--var-file", varFile)
	}
	args = append(args, t.ExtraArgs...)
	args = append(args, ".")
	// #nosec G204
	c := exec.Command(d.GetExePath("opal"), args...)
	c.Dir = t.GetDirectory()
	c.Stderr = os.Stderr
	exec := t.ExecuteCommand(c)
	result.ExecuteResult = exec
	if !exec.ExpectExitCode(0, 1) {
		return result, nil
	}
	n, ok := exec.ParseJSON()
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
		})
	}
}
