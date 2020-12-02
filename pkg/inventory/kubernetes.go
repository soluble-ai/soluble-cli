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
