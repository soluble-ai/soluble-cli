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
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type kubernetesDetector int

var _ FileDetector = kubernetesDetector(0)

func (kubernetesDetector) DetectFileName(m *Manifest, path string) ContentDetector {
	if filepath.Base(path) == "kustomization.yaml" {
		m.KustomizeDirectories.Add(filepath.Dir(path))
		return nil
	}
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") ||
		strings.HasSuffix(path, ".json") {
		return kubernetesDetector(0)
	}
	return nil
}

func (kubernetesDetector) DetectContent(m *Manifest, path string, content []byte) {
	d := decodeDocument(path, content)
	if filepath.Base(path) == "Chart.yaml" && d["apiVersion"] != "" {
		// Looks like a helm chart
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
		// ignore templates in kustomize directories
		if m.KustomizeDirectories.Contains(filepath.Dir(path)) {
			return
		}
		m.KubernetesManifestDirectories.Add(filepath.Dir(path))
	}
}

func (kubernetesDetector) FinalizeDetection(m *Manifest) {
	// We want to remove helm subcharts, which are helm
	// charts in a subdirectory under "charts/" under another helm
	// chart
	charts := util.NewStringSet()
	chartPaths := m.HelmCharts.Values()
	sort.Strings(chartPaths)
	for _, chart := range chartPaths {
		parts := strings.Split(chart, string(os.PathSeparator))
		n := len(parts)
		if n > 2 && parts[n-2] == "charts" {
			parentChart := filepath.Join(parts[0 : n-2]...)
			if m.HelmCharts.Contains(parentChart) {
				continue
			}
		}
		charts.Add(chart)
	}
	m.HelmCharts = *charts
}
