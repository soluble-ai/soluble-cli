package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/inventory/terraformsettings"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

func RunTerraformGet(dir string, opts tools.RunOpts) error {
	inv := inventory.Do(dir)
	for _, rootModule := range inv.TerraformModules.Values() {
		dir := filepath.Join(dir, rootModule)
		terraformExe, err := downloadTerraformExe(dir)
		if err != nil {
			return err
		}
		cmd := exec.Command(terraformExe, "get")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Dir = dir
		opts.LogCommand(cmd)
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func RunTerraformInit(dir string, opts tools.RunOpts, tfCommand string) (*terraformInit, error) {
	tfi := &terraformInit{}
	inv := inventory.Do(dir)
	for _, rootModule := range inv.TerraformRootModules.Values() {
		dir := filepath.Join(dir, rootModule)
		var terraformArgs []string
		if tfCommand != "" {
			terraformArgs = strings.Split(tfCommand, " ")
		} else {
			terraformExe, err := downloadTerraformExe(dir)
			if err != nil {
				return nil, err
			}
			terraformArgs = []string{terraformExe}
		}
		tfi.files = append(tfi.files, newDeletedFile(filepath.Join(dir, ".terraform", "terraform.tfstate")))
		terraformArgs = append(terraformArgs, "init", "-backend=false")
		// #nosec G204
		cmd := exec.Command(terraformArgs[0], terraformArgs[1:]...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Dir = dir
		opts.LogCommand(cmd)
		err := cmd.Run()
		if err != nil {
			tfi.Restore()
			return nil, err
		}
	}
	return tfi, nil
}

func downloadTerraformExe(dir string) (string, error) {
	settings := terraformsettings.Read(dir)
	installer := &tools.RunOpts{}
	d, err := installer.InstallTool(&download.Spec{
		Name:             "terraform",
		RequestedVersion: settings.GetTerraformVersion(),
	})
	if err != nil {
		return "", err
	}
	return d.GetExePath("terraform"), nil
}

type deletedFile struct {
	mode os.FileMode
	path string
	data []byte
}

func newDeletedFile(path string) *deletedFile {
	f := &deletedFile{
		path: path,
	}
	fi, err := os.Stat(path)
	if err == nil {
		f.mode = fi.Mode()
		f.data, _ = os.ReadFile(path)
		if err := os.Remove(f.path); err != nil {
			log.Warnf("Could not remove {warning:%s} - {danger:%s}", f.path, err)
			f.data = nil
		} else {
			log.Infof("Temporarily removing {info:%s}", f.path)
		}
	}
	return f
}

func (f *deletedFile) restore() {
	if f.data != nil {
		err := os.WriteFile(f.path, f.data, f.mode)
		if err != nil {
			log.Warnf("Could not restore {info:%s} - {danger:%s}", f.path, err)
		} else {
			log.Infof("Restored {info:%s}", f.path)
		}
		f.data = nil
	}
}

type terraformInit struct {
	files []*deletedFile
}

func (tfi *terraformInit) Restore() {
	for _, sf := range tfi.files {
		sf.restore()
	}
}
