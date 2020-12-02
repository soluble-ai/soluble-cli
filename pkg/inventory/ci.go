package inventory

import (
	"path/filepath"
	"strings"
)

type cidetector int

var _ FileDetector = cidetector(0)

func (cidetector) DetectDirName(m *Manifest, path string) {
	switch path {
	case ".buildkite":
		m.CISystems.Add("buildkite")
	case ".circleci":
		m.CISystems.Add("circleci")
	}
}

func (cidetector) DetectFileName(m *Manifest, path string) ContentDetector {
	switch path {
	case "Jenkinsfile":
		m.CISystems.Add("jenkins")
	case "azure-pipelines.yml":
		m.CISystems.Add("azure")
	case ".travis.yml":
		m.CISystems.Add("travis")
	case ".drone.yml":
		m.CISystems.Add("drone")
	case ".gitlab-ci.yml":
		m.CISystems.Add("gitlab")
	}
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		if filepath.Dir(path) == filepath.Join(".github", "workflows") {
			m.CISystems.Add("github")
		}
	}
	return nil
}
