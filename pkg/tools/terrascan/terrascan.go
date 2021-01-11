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
	"github.com/soluble-ai/soluble-cli/pkg/util"
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
	d, err := t.InstallTool(&download.Spec{
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
	if err != nil && util.ExitCode(err) != 3 {
		// terrascan exits with exit code 3 if violations were found
		return nil, err
	}
	n, err := jnode.FromJSON(output)
	if err != nil {
		return nil, err
	}
	for _, result := range n.Path("results").Elements() {
		violations := result.Path("violations")
		if violations.Size() > 0 {
			result.Put("violations", util.RemoveJNodeElementsIf(violations, func(e *jnode.Node) bool {
				return t.IsExcluded(e.Path("file").AsText())
			}))
		}
	}
	result := &tools.Result{
		Data:         n,
		Directory:    t.GetDirectory(),
		PrintPath:    []string{"results", "violations"},
		PrintColumns: []string{"category", "severity", "file", "line", "rule_id", "description"},
	}
	if d.Version != "" {
		result.AddValue("TERRASCAN_VERSION", d.Version)
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
