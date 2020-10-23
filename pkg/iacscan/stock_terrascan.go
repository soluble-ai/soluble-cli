package iacscan

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/client"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

var _ IacScanner = &StockTerrascan{}

const (
	policyZip = "rego-policies.zip"
	rulesPath = "terrascan"
)

type StockTerrascan struct {
	directory  string
	apiClient  client.Interface
	policyPath string
}

func (t *StockTerrascan) Run() (*Result, error) {
	m := download.NewManager()
	d, err := m.InstallGithubRelease("accurics", "terrascan", "")
	if err != nil {
		return nil, err
	}

	if err = t.downloadPolicies(); err != nil {
		return nil, err
	}

	log.Infof("Running {info:terrascan} -d %s", t.directory)
	program := filepath.Join(d.Dir, "terrascan")
	// the -t argument is required but it only selects what policies are
	// selected if the -p option isn't used.  Since we're using -p,
	// we can pass any valid value.
	scan := exec.Command(program, "scan", "-t", "aws", "-d", t.directory, "-p", t.policyPath, "-o", "json")
	scan.Stderr = os.Stderr
	scan.Stderr = os.Stderr
	output, err := scan.Output()
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		// terrascan exits with exit code 3 if violations were found
		if !ok || ee.ExitCode() != 3 {
			return nil, err
		}
	}
	n, err := jnode.FromJSON(output)
	if err != nil {
		return nil, err
	}
	return &Result{
		N:            n,
		PrintPath:    []string{"results", "violations"},
		PrintColumns: []string{"category", "severity", "file", "line", "rule_id", "description"},
	}, nil
}

func (t *StockTerrascan) downloadPolicies() error {
	m := download.NewManager()
	url := fmt.Sprintf("%s/api/v1/org/%s/opa/%s", t.apiClient.GetClient().HostURL, t.apiClient.GetOrganization(),
		policyZip)
	d, err :=
		m.Install("terrascan-policies", "latest", url, download.WithBearerToken(t.apiClient.GetClient().Token))
	if err != nil {
		return err
	}
	t.policyPath = filepath.Join(d.Dir, rulesPath)
	return nil
}
