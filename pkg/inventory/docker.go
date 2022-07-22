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

type dockerDetector int

var _ FileDetector = dockerDetector(0)

func (d dockerDetector) DetectFileName(m *Manifest, path string) ContentDetector {
	base := strings.ToLower(filepath.Base(path))
	if base == "dockerfile" || strings.HasPrefix(base, "dockerfile.") || strings.HasSuffix(base, ".dockerfile") {
		return d
	}
	return nil
}

func (dockerDetector) DetectContent(m *Manifest, path string, content []byte) {
	if strings.Contains(string(content), "FROM ") {
		m.DockerDirectories.Add(filepath.Dir(path))
		m.Dockerfiles.Add(path)
	}
}
