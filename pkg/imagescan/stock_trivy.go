package imagescan

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

const (
	outputFile = "vulnerabilities.json"
)

type StockTrivy struct {
	image string
	name  string
}

func (t *StockTrivy) Run() (*Result, error) {
	m := download.NewManager()
	d, err := m.InstallGithubRelease("aquasecurity", trivy, "")
	if err != nil {
		return nil, err
	}
	program := filepath.Join(d.Dir, trivy)
	scan := exec.Command(program, "image", "-f", "json", "-o", outputFile, t.image)
	scan.Stderr = os.Stderr
	output, err := scan.Output()
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		// exit code 0 means vulnerabilities are found
		if !ok || ee.ExitCode() != 0 {
			return nil, err
		}
	}
	// print trivy messages to std out
	log.Infof(string(output))

	_, err = os.Stat(outputFile)
	if err != nil {
		log.Warnf("No vulnerability report found")
		return nil, err
	}

	// read results from the report
	results, err := os.Open(outputFile)
	if err != nil {
		return nil, err
	}
	defer results.Close()
	byteValue, _ := ioutil.ReadAll(results)

	// clean up json format

	n, err := jnode.FromJSON(byteValue)
	if err != nil {
		return nil, err
	}

	// delete the report
	_ = os.Remove(outputFile)

	return &Result{
		N:            n.Get(0),
		PrintPath:    []string{"Vulnerabilities"},
		PrintColumns: []string{"PkgName", "VulnerabilityID", "Severity", "InstalledVersion", "FixedVersion", "Title"},
	}, nil
}

// Name of the image scanning module
func (t *StockTrivy) Name() string {
	return t.name
}
