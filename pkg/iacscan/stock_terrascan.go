package iacscan

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/client"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"gopkg.in/yaml.v2"
)

var _ IacScanner = &StockTerrascan{}
var supportedTypes [4]string = [4]string{"aws", "gcp", "azure", "k8s"}

const (
	policyZip = "rego-policies.zip"
	rulesPath = "metadata-opa-policies/policies/accurics/terrascan"
)

type StockTerrascan struct {
	Directory  string
	policyPath string
	Report     bool
	apiClient  client.Interface
}

func (t *StockTerrascan) Run() (*jnode.Node, error) {
	opts := options.ClientOpts{}
	t.apiClient = opts.GetAPIClient()

	m := download.NewManager()
	d, err := m.InstallGithubRelease("accurics", "terrascan", "")
	if err != nil {
		return nil, err
	}

	if err = t.downloadPolicies(m.DownloadDir); err != nil {
		return nil, err
	}

	program := filepath.Join(d.Dir, "terrascan")
	init := exec.Command(program, "init")
	init.Stdout = os.Stderr
	init.Stderr = os.Stderr
	log.Infof("%s", init.Args)
	if err = init.Run(); err != nil {
		return nil, fmt.Errorf("terrascan init failed: %w", err)
	}

	v := []map[string]interface{}{}
	for _, iacType := range supportedTypes {
		scan := exec.Command(program, "scan", "-t", iacType, "-d", t.Directory, "-p", t.policyPath)
		out := &bytes.Buffer{}
		scan.Stderr = os.Stderr
		scan.Stdout = out
		scan.Stderr = os.Stderr
		log.Infof("%s", scan.Args)
		_ = scan.Run()

		var output map[string]interface{}
		if err = yaml.Unmarshal(out.Bytes(), &output); err != nil {
			return nil, fmt.Errorf("could not parse terrascan output: %w", err)
		}
		v = append(v, output)
	}

	result := mergeViolationResults(v...)
	if t.Report {
		file, err := os.Create("results.json")
		if err != nil {
			return jnode.FromMap(result), err
		}
		defer file.Close()

		j, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return jnode.FromMap(result), err
		}

		w := bufio.NewWriter(file)
		_, err = w.WriteString(string(j))
		w.Flush()

		dir, _ := os.Getwd()
		files := []string{filepath.Join(dir, file.Name())}

		ciOptions := []client.Option{client.XCPWithCIEnv}
		err = t.apiClient.XCPPost(opts.GetOrganization(), "terrascan", files, nil, ciOptions...)
		if err != nil {
			return jnode.FromMap(result), err
		}
	}
	return jnode.FromMap(result), nil
}

func (t *StockTerrascan) downloadPolicies(downloadDir string) error {
	policyPath := filepath.Join(downloadDir, rulesPath)
	_, err := os.Stat(policyPath)

	// TODO: Check the version or tag to determine the download rather than checking directory
	if os.IsNotExist(err) {
		// Download the policies from the API server to the specified policyPath
		path := fmt.Sprintf("org/{org}/opa/%s", policyZip)
		t.apiClient.GetClient().SetOutputDirectory(downloadDir)
		_, err := t.apiClient.Get(path, func(req *resty.Request) {
			req.SetOutput(policyZip)
		})
		if err != nil {
			return err
		}

		policiesZipPath := filepath.Join(downloadDir, policyZip)
		cmd := exec.Command("unzip", "-o", "-d", downloadDir, policiesZipPath)
		if err = cmd.Run(); err != nil {
			return err
		}

		// remove the zip file
		_ = os.Remove(policiesZipPath)
	}

	t.policyPath = policyPath
	return nil
}

func mergeViolationResults(maps ...map[string]interface{}) map[string]interface{} {
	var lowCount, highCount, mediumCount, totalCount int
	violationsStats := make(map[string]int)

	var violations []interface{}
	for _, m := range maps {
		for _, violationType := range m["results"].(map[interface{}]interface{})["violations"].([]interface{}) {
			violations = append(violations, violationType)
			severity := violationType.(map[interface{}]interface{})["severity"].(string)
			if strings.ToLower(severity) == "high" {
				highCount++
			} else if strings.ToLower(severity) == "medium" {
				mediumCount++
			} else if strings.ToLower(severity) == "low" {
				lowCount++
			}
			totalCount++
		}
	}
	violationsStats["total"] = totalCount
	violationsStats["low"] = lowCount
	violationsStats["medium"] = mediumCount
	violationsStats["high"] = highCount

	output := make(map[string]interface{})
	output["count"] = violationsStats
	output["violations"] = violations

	result := make(map[string]interface{})
	result["results"] = output
	return result
}
