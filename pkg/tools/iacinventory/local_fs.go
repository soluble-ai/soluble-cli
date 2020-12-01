package iacinventory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/cmd/iacscan"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

var _ tools.Interface = &FSIACInventoryScanner{}

type FSIACInventoryScanner struct {
	tools.ToolOpts
	Directories []string
	IACScan     bool
}

func (f *FSIACInventoryScanner) Name() string {
	return "filesystem-iac-inventory"
}

func (f *FSIACInventoryScanner) Register(c *cobra.Command) {
	f.ToolOpts.Register(c)
	flags := c.Flags()
	flags.StringSliceVar(&f.Directories, "dir", nil, "Local directories to scan. May be repeated.")
	flags.BoolVar(&f.IACScan, "iac-scan", false, "Scan the results with the Soluble iac-scan command. Results are automatically uploaded to the Soluble API.")
}

func (f *FSIACInventoryScanner) Run() (*tools.Result, error) {
	if f.Directories == nil {
		return nil, fmt.Errorf("no directories (--dir) provided for scan")
	}
	var err error
	result := &tools.Result{
		Data: jnode.NewObjectNode(),
		// Values?
		Values:    map[string]string{"DIRECTORIES": f.Directories[0]}, // BUG
		PrintPath: []string{"directories"},
		PrintColumns: []string{
			"directory", "ci_systems", "terraform_dir_count", "cloudformation_dir_count", "dockerfile_count", "k8s_manifest_dir_count",
		},
	}
	a := result.Data.PutArray("directories")
	for _, dir := range f.Directories {
		iacRes, err := Directory(dir)
		if err != nil {
			return nil, err
		}
		dat, err := json.Marshal(iacRes)
		if err != nil {
			return nil, err
		}
		r, err := jnode.FromJSON(dat)
		if err != nil {
			return nil, err
		}
		r.Put("terraform_dir_count", r.Path("terraform_dirs").Size())
		r.Put("cloudformation_dir_count", r.Path("cloudformation_dirs").Size())
		r.Put("dockerfile_count", r.Path("dockerfile_files").Size())
		r.Put("k8s_manifest_dir_count", r.Path("k8s_manifest_dirs").Size())
		a.Append(r)

		if !f.IACScan {
			continue
		}

		scans := make(map[string][]string)
		// the toolName_dirs syntax specified here w/ _dirs because we may eventually need _files (cfnguard)
		scans["terraform_dirs"] = iacRes.TerraformDirs
		scans["cloudformation_dirs"] = iacRes.CloudformationDirs
		scans["k8s_dirs"] = iacRes.K8sManifestDirs

		wg := sync.WaitGroup{}
		for iacType, sDirs := range scans {
			switch iacType {
			case "terraform_dirs":
				for _, sDir := range sDirs {
					wg.Add(1)
					// go scan("terrascan", filepath.Join(dir, sDir), &wg)
					go scan("checkov", filepath.Join(dir, sDir), &wg)
					// go scan("tfsec", filepath.Join(dir, sDir), &wg)
					time.Sleep(1 * time.Second)
				}
			case "cloudformation_dirs":
				for _, sDir := range sDirs {
					wg.Add(1)
					go scan("cfn-python-lint", filepath.Join(dir, sDir), &wg)
					time.Sleep(1 * time.Second)
				}
			}
		}
		wg.Wait()
		log.Infof("Scan complete.")
	}
	return result, err
}

// scan is intended to be run in a goroutine. Errors are ignored, and all results
// are unconditionally uploaded.
func scan(toolName, dir string, wg *sync.WaitGroup) {
	defer wg.Done()
	cmd := iacscan.Command()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{toolName, "--directory", dir, "--upload"})
	if err := cmd.Execute(); err != nil {
		log.Errorf("error running iac-scan from inventory dir: %q:\n%w", dir, err)
	}
}

// dedupe removes "duplicates" for tools, meaning sub-directories that have/will be covered
// by scans performed on parent directories (by recursive tools).
// For example:
//   []string{"path/decended/to/subdir", "path"}
// becomes...
//   []string{"path"}
//
// This is useful for tools that are themselves recursive.
func dedupe(scans map[string][]string) map[string][]string {
	deduped := make(map[string][]string)
	for tool, dirs := range scans {
		sort.Strings(dirs) // directories must be sorted shortest to longest
		deduped[tool] = []string{}
		for _, dir := range dirs {
			contains := false
			for _, dedupedDir := range deduped[tool] {
				if strings.Contains(dir, dedupedDir) {
					contains = true
				}
			}
			if !contains {
				deduped[tool] = append(deduped[tool], dir)
			}
		}
	}
	return deduped
}
