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
	}
}
