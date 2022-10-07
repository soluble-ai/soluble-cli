package opal

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/tools/util"
)

func (t *Tool) runTerraformGet() error {
	inv := inventory.Do(t.GetDirectory())
	for _, rootModule := range inv.TerraformModules.Values() {
		dir := filepath.Join(t.GetDirectory(), rootModule)

		terraformExe, err := util.DownloadTerraformExe(dir)
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
