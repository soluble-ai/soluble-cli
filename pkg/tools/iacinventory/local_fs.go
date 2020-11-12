package iacinventory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"

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
		fmt.Println()
		cmd := iacscan.Command()
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		log.Infof("Scanning Terraform Directories...")
		for _, sDir := range iacRes.TerraformDirs {
			cmd.SetArgs([]string{"terrascan", "--directory", filepath.Join(dir, sDir), "--upload"})
			if err := cmd.Execute(); err != nil {
				return nil, fmt.Errorf("error running iac-scan from inventory dir %q:\n%w", dir, err)
			}
			cmd.SetArgs([]string{"checkov", "--directory", filepath.Join(dir, sDir), "--upload"})
			if err := cmd.Execute(); err != nil {
				return nil, fmt.Errorf("error running iac-scan from inventory dir %q:\n%w", dir, err)
			}
			cmd.SetArgs([]string{"tfsec", "--directory", filepath.Join(dir, sDir), "--upload"})
			if err := cmd.Execute(); err != nil {
				return nil, fmt.Errorf("error running iac-scan from inventory dir %q:\n%w", dir, err)
			}
			log.Debugf("scanned %s", filepath.Join(dir, sDir))
			// fmt.Println(buf.Bytes()) // we may want to do something with the output at some point
			buf.Reset()
		}
		log.Infof("Scanning Cloudformation Directories...")
		for _, sDir := range iacRes.CloudformationDirs {
			cmd.SetArgs([]string{"cfn-python-lint", "--directory", filepath.Join(dir, sDir), "--upload"})
			if err := cmd.Execute(); err != nil {
				return nil, fmt.Errorf("error running iac-scan from inventory dir %q:\n%w", dir, err)
			}
			log.Debugf("scanned %s", filepath.Join(dir, sDir))
			// fmt.Println(buf.Bytes()) // we may want to do something with the output at some point
			buf.Reset()
		}
		// we also have to loop cloudformation files, because #cfnguard
		/*
			for _, sFile := range iacRes.CloudformationFiles {
				cmd.SetArgs([]string{"cloudformationguard", "--file", filepath.Join(dir, sFile)}, "--upload")
				if err := cmd.Execute(); err != nil {
					return nil, fmt.Errorf("error running iac-scan from inventory dir %q:\n%w", dir, err)
				}
				log.Debugf("scanned %s", filepath.Join(dir, sFile))
				// fmt.Println(buf.Bytes()) // we may want to do something with the output at some point
				buf.Reset()
			}
		*/
		log.Infof("Scan complete.")
	}
	return result, err
}
