package cloudformationguard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/client"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

const (
	rulesZip = "cfn-guard.zip"
)

type Tool struct {
	File      string
	APIClient client.Interface

	rulesInstallDir string
	program         string
	version         string
}

var _ tools.RunsWithAPIClient = &Tool{}

func (t *Tool) Name() string {
	return "cloudformationguard"
}

func (t *Tool) SetAPIClient(apiClient client.Interface) {
	t.APIClient = apiClient
}

func (t *Tool) Run() (*tools.Result, error) {
	if err := t.installProgram(); err != nil {
		return nil, fmt.Errorf("could not download cloudformation-guard: %w", err)
	}
	if err := t.downloadRules(); err != nil {
		return nil, fmt.Errorf("error downloading cloudformation-guard rules: %w", err)
	}
	// #nosec G204
	scan := exec.Command(t.program, "check", "--template", t.File, "--rule_set",
		fmt.Sprintf("%s/cfn-guard/security.ruleset", t.rulesInstallDir))
	scan.Stderr = os.Stderr
	log.Infof("Running {info:%s}", strings.Join(scan.Args, " "))
	output, err := scan.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			if ee.ExitCode() != 2 {
				return nil, err
			}
		}
	}
	n := jnode.NewObjectNode()
	s := string(output)
	n.Put("raw_output", s)
	failures := n.PutArray("failures")
	for _, f := range parseFailures(s) {
		fn, _ := print.ToResult(f)
		failures.Append(fn)
	}
	result := &tools.Result{
		Data: n,
		Values: map[string]string{
			"CLOUDFORMATION_GUARD_VERSION": t.version,
		},
		Directory:    filepath.Dir(t.File),
		PrintPath:    []string{"failures"},
		PrintColumns: []string{"resource", "attribute", "attribute_value", "message"},
	}
	result.AddFile(t.File)
	return result, nil
}

func (t *Tool) installProgram() error {
	m := download.NewManager()
	d, err := m.Install(&download.Spec{
		URL:                  "github.com/aws-cloudformation/cloudformation-guard",
		GithubReleaseMatcher: func(release string) bool { return download.IsMatchingOS(release, runtime.GOOS) },
	})
	if err != nil {
		return fmt.Errorf("error downloading cloudformation-guard from GitHub: %w", err)
	}
	binDir := filepath.Join(d.Dir, "cfn-guard-osx")
	t.program = binDir + "/cfn-guard"
	t.version = d.Version
	return nil
}

func (t *Tool) downloadRules() error {
	d, err := tools.InstallAPIServerArtifact(t.APIClient, "cfn-guard-policies",
		fmt.Sprintf("/api/v1/org/{org}/cfn-guard/%s", rulesZip))
	if err != nil {
		return err
	}
	t.rulesInstallDir = d.Dir
	return nil
}
