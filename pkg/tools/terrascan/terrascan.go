package terrascan

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

var (
	_ tools.Interface = &Tool{}
)

const (
	policyZip = "rego-policies.zip"
	rulesPath = "terrascan"
)

type Tool struct {
	tools.DirectoryBasedToolOpts

	policyPath string
}

func (t *Tool) Name() string {
	return "terrascan"
}

func (t *Tool) Run() (*tools.Result, error) {
	m := download.NewManager()
	d, err := m.Install(&download.Spec{
		URL: "github.com/accurics/terrascan",
	})
	if err != nil {
		return nil, err
	}
	if err = t.downloadPolicies(); err != nil {
		return nil, err
	}
	log.Infof("Running {info:terrascan} -d %s", t.GetDirectory())
	program := filepath.Join(d.Dir, "terrascan")
	// the -t argument is required but it only selects what policies are
	// selected if the -p option isn't used.  Since we're using -p,
	// we can pass any valid value.
	scan := exec.Command(program, "scan", "-t", "aws", "-d", t.GetDirectory(), "-p", t.policyPath, "-o", "json")
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
	result := &tools.Result{
		Data:      n,
		Directory: t.GetDirectory(),
		Values: map[string]string{
			"TERRASCAN_VERSION": d.Version,
		},
		PrintPath:    []string{"results", "violations"},
		PrintColumns: []string{"category", "severity", "file", "line", "rule_id", "description"},
	}
	for _, v := range n.Path("results").Path("violations").Elements() {
		file := v.Path("file").AsText()
		if file != "" {
			result.AddFile(file)
		}
	}
	return result, nil
}

func (t *Tool) downloadPolicies() error {
	d, err := t.InstallAPIServerArtifact("terrascan-policies",
		fmt.Sprintf("/api/v1/org/{org}/opa/%s", policyZip))
	if err != nil {
		return err
	}
	t.policyPath = filepath.Join(d.Dir, rulesPath)
	return nil
}
