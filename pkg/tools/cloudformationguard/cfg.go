package cloudformationguard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/client"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/iacinventory"
)

const (
	rulesZip  = "cloudformation-guard-rulesets.zip"
	rulesPath = "cloudformation-guard"
)

type Tool struct {
	Directory string
	APIClient client.Interface

	rulesPath string
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
	binDir := filepath.Join(d.Dir, "cfn-guard-osx") // TODO: directory varies by name. InstallGithubRelease should extract to consistent dirname.
	program := binDir + "/cfn-guard"

	if err := t.downloadRules(); err != nil {
		return nil, fmt.Errorf("error downloading cloudformation-guard rules: %w", err)
	}

	var ruleFiles []string
	err = filepath.Walk(t.rulesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if filepath.Ext(path) != "ruleset" {
			return nil
		}
		ruleFiles = append(ruleFiles, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error getting rule files: %w", err)
	}

	// TODO: move this out of inline func
	type violation struct {
		OffendingFileRel string `json:"file"`
		OffendingFileAbs string `json:"file_abs"`
		RuleFile         string `json:"rule_file"`  // the rule file that matched the violation
		RawOutput        string `json:"raw_output"` // TODO: parse into fields
	}
	var violations []violation
	err = filepath.Walk(t.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !iacinventory.IsCloudFormationFile(path, info) {
			return nil
		}
		// handle relative paths for template files
		templateAbsPath := path
		if !filepath.IsAbs(templateAbsPath) {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			templateAbsPath, err = filepath.Abs(filepath.Join(wd, path))
			if err != nil {
				return err
			}
		}
		log.Debugf("checking file %q with cfn-guard", templateAbsPath)
		for _, rf := range ruleFiles {
			scan := exec.Command(program, "check", "--template", templateAbsPath, "--rule_set", rf)
			scan.Stderr = os.Stdout
			output, err := scan.Output()
			if err != nil {
				ee, ok := err.(*exec.ExitError)
				// TODO: handle exit code
				if !ok || ee.ExitCode() != 0 {
					log.Infof("exit code: %v", ee.ExitCode())
				}
			}
			violations = append(violations, violation{
				OffendingFileAbs: templateAbsPath,
				OffendingFileRel: path,
				RuleFile:         rf,
				RawOutput:        string(output),
			})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error scanning repo with cfn-guard: %w", err)
	}

	// Eventually this will need a parser, but inline is fine for PoC
	result := &tools.Result{
		Data:         jnode.FromSlice(violations),
		Directory:    t.Directory,
		PrintPath:    []string{""},
		PrintColumns: []string{"file", "result"},
	}
	// add offending files to output
	for _, v := range violations {
		if v.RawOutput == "" {
			continue
		}
		result.AddFile(v.OffendingFileAbs)
	}
	return result, nil
}

func (t *Tool) downloadRules() error {
	m := download.NewManager()
	url := fmt.Sprintf("%s/api/v1/org/%s/cfn-guard/%s", t.APIClient.GetClient().HostURL, t.APIClient.GetOrganization(), rulesZip)
	d, err := m.Install("cfn-guard-policies", "latest", url, download.WithBearerToken(t.APIClient.GetClient().Token))
	if err != nil {
		return err
	}
	t.rulesPath = filepath.Join(d.Dir, rulesPath)
	return nil
}
