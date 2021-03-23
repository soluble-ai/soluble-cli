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

type kubernetesDetector int

var _ FileDetector = kubernetesDetector(0)

func (kubernetesDetector) DetectFileName(m *Manifest, path string) ContentDetector {
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") ||
		strings.HasSuffix(path, ".json") {
		return kubernetesDetector(0)
	}
	return nil
}

func (kubernetesDetector) DetectContent(m *Manifest, path string, content []byte) {
	d := PartialDecode(path, content)
	if filepath.Base(path) == "Chart.yaml" && d["apiVersion"] != "" {
		// assume this is a helm chart
		m.HelmCharts.Add(filepath.Dir(path))
		return
	}
	if d["apiVersion"] != "" && d["kind"] != "" {
		// ignore templates in helm charts
		for _, ch := range m.HelmCharts.Values() {
			if strings.HasPrefix(path, ch) {
				return
			}
		}
		m.KubernetesManifestDirectories.Add(filepath.Dir(path))
	}
}
