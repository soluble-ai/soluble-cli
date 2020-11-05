package iacinventory

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type Repo interface {
	SetTerraformDirs([]string)
	SetCloudformationDirs([]string)
	SetDockerfileDirs([]string)
	SetK8sManifestDirs([]string)
	SetCISystems([]string)
}

// scanRepo for IaC files
func scanRepo(r Repo, dir string) error {
	terraformDirs := util.NewStringSet()
	cfnDirs := util.NewStringSet()
	dockerFileDirs := util.NewStringSet()
	k8sManifestDirs := util.NewStringSet()
	ciSystems := util.NewStringSet()
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		ci, err := walkCI(path, info, err)
		if err != nil {
			return err
		}
		if ci != "" {
			ciSystems.Add(string(ci))
		}
		if info.Mode().IsRegular() {
			dirName, _ := filepath.Rel(dir, filepath.Dir(path))
			splitPath := strings.SplitN(dirName, string(filepath.Separator), 2)
			pathRelativeToRepoRoot := "."
			if len(splitPath) > 1 {
				// if we're not in the root, set the directory
				// relative to the git repository root
				pathRelativeToRepoRoot = splitPath[1]
			}
			if isTerraformFile(path, info) {
				terraformDirs.Add(pathRelativeToRepoRoot)
			}
			if isCloudFormationFile(path, info) {
				cfnDirs.Add(pathRelativeToRepoRoot)
			}
			if isDockerFile(path, info) {
				dockerFileDirs.Add(pathRelativeToRepoRoot)
			}
			if isKubernetesManifest(path, info) {
				k8sManifestDirs.Add(pathRelativeToRepoRoot)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	r.SetTerraformDirs(terraformDirs.Values())
	r.SetCloudformationDirs(cfnDirs.Values())
	r.SetDockerfileDirs(dockerFileDirs.Values())
	r.SetK8sManifestDirs(k8sManifestDirs.Values())
	r.SetCISystems(ciSystems.Values())
	return nil
}
