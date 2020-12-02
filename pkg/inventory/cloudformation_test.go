package inventory

import "testing"

func TestCloudformationDetector(t *testing.T) {
	var testCases = []struct {
		name, content string
		match         bool
	}{
		{"foo.yaml", `---
AWSTemplateFormatVersion: '2010-09-09'`, true},
		{"foo.yaml", "#AWSTemplateFormatVersion: '2010-09-09", false},
		{"foo.json", `{ "AWSTemplateFormatVersion" :
		"2010-09-09", "bar": 1`, true},
	}
	d := cloudformationDetector(0)
	for _, tc := range testCases {
		m := &Manifest{}
		d.DetectContent(m, tc.name, []byte(tc.content))
		if tc.match && (m.CloudformationFiles.Len() != 1 || m.CloudformationFiles.Get(0) != tc.name) {
			t.Error(tc)
		} else if !tc.match && m.CloudformationFiles.Len() != 0 {
			t.Error(tc)
		}
	}
}
