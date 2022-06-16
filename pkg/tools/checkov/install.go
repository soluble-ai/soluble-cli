package checkov

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

func installAndAddToPath(url, executable, defaultVersion string) error {
	installer := &tools.RunOpts{}
	d, err := installer.InstallTool(&download.Spec{
		URL:              url,
		RequestedVersion: defaultVersion,
	})
	if err != nil {
		return err
	}
	dir := filepath.Dir(d.GetExePath(executable))
	// add to path
	path := os.Getenv("PATH")
	if path == "" {
		path = dir
	} else {
		path = fmt.Sprintf("%s%c%s", path, os.PathListSeparator, dir)
	}
	log.Infof("Adding {info:%s} to PATH", dir)
	os.Setenv("PATH", path)
	return nil
}
