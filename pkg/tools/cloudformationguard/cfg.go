package cloudformationguard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/client"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

type Tool struct {
	CloudformationFiles []string
	CloudformationRules []string // policy download TBD
	APIClient           client.Interface

	policyPath string
}

func (t *Tool) Name() string {
	return "cloudformationguard"
}

func (t *Tool) Run() (*tools.Result, error) {
	m := download.NewManager()
	d, err := m.InstallGithubRelease("aws-cloudformation", "cloudformation-guard", "")
	if err != nil {
		return nil, fmt.Errorf("error downloading cloudformation-guard from GitHub: %w", err)
	}
	filepath.Join(d.Dir, "cloudformation-guard")
	scan := exec.Command(program, "check", "--template", templateFile, "--rule_set", ruleFile)
	scan.Stderr = os.Stdout
	output, err := scan.Output()
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		// TODO: handle exit code
		if !ok || ee.ExitCode() != 0 {
			log.Infof("exit code: %v", ee.ExitCode())
		}
	}
	// TODO: jnode?
	// n, err := jnode.FromJSON(output)
	log.Infof("cfn-guard output: %v", string(output))

	result := &tool.Result{
		Data: output, // TODO

	}
	log.Infof("Running {info:cfn-guard} check")
}
