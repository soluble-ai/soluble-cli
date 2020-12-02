package inventory

import (
	"strings"
)

type cloudformationDetector int

var _ FileDetector = cloudformationDetector(0)

func (cloudformationDetector) DetectFileName(m *Manifest, path string) ContentDetector {
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".json") {
		return cloudformationDetector(0)
	}
	return nil
}

func (cloudformationDetector) DetectContent(m *Manifest, path string, buf []byte) {
	d := PartialDecode(path, buf)
	if _, ok := d["AWSTemplateFormatVersion"]; ok {
		m.CloudformationFiles.Add(path)
	}
}
