package iacinventory

import (
	"os"
	"path/filepath"
	"strings"
)

// CI is the type of CI used for a repository.
type CI string

const (
	CIBuildkite CI = "buildkite"
	CIGithub    CI = "github"
	CIGitlab    CI = "gitlab"
	CICircle    CI = "circle"
	CIDrone     CI = "drone"
	CIJenkins   CI = "jenkins"
	CIAzure     CI = "azure"
	CITravis    CI = "travis"
)

// walkCI Tests a given CI system against a file for a match, implements filepath.WalkFunc
func walkCI(path string, info os.FileInfo, err error) (CI, error) {
	if err != nil {
		return "", err
	}
	// Find if the repository uses a CI system.
	// TODO: check for jenkins, which uses file not directory
	var ci CI
	if info.IsDir() {
		switch info.Name() {
		case "workflows":
			if strings.HasSuffix(path, filepath.Join(".github", "workflows")) {
				ci = CIGithub
			}
		case ".buildkite":
			ci = CIBuildkite
		case ".gitlab":
			ci = CIGitlab
		case ".drone":
			ci = CIDrone
		case ".circleci":
			ci = CICircle
		}
	}
	if info.Mode().IsRegular() {
		switch info.Name() {
		case "Jenkinsfile":
			ci = CIJenkins
		case "azure-pipelines.yml":
			ci = CIAzure
		case ".travis.yml":
			ci = CITravis
		}
	}
	return ci, nil
}
