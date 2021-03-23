// Copyright 2021 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
