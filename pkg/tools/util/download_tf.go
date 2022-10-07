package util

import (
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/inventory/terraformsettings"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

func DownloadTerraformExe(dir string) (string, error) {
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
