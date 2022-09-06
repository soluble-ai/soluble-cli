package opal

import (
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/inventory/terraformsettings"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"os"
	"os/exec"
	"path/filepath"
)

func (t *Tool) runTerraformGet() (error) {
	inv := inventory.Do(t.GetDirectory())
	for _, rootModule := range inv.TerraformModules.Values() {
		dir := filepath.Join(t.GetDirectory(), rootModule)

		terraformExe, err := t.downloadTerraformExe(dir)
		if err != nil {
			return err
		}
		cmd := exec.Command(terraformExe, "get")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Dir = dir
		t.LogCommand(cmd)
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}


func (t *Tool) downloadTerraformExe(dir string) (string, error) {
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