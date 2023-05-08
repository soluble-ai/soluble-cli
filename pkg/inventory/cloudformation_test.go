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

import "testing"

func TestCloudformationDetector(t *testing.T) {
	var testCases = []struct {
		name, content string
		match         bool
	}{
		{"foo.yaml", `---
AWSTemplateFormatVersion: '2010-09-09'`, true},
		{"foo.yml", `---
AWSTemplateFormatVersion: '2010-09-09'`, true},
		{"foo.yml", `---
AWSTemplateFormatVersion: 2010-09-09`, true},
		{"foo.yaml", "#AWSTemplateFormatVersion: '2010-09-09'", false},
		{"foo.json", `{ "AWSTemplateFormatVersion" :
		"2010-09-09", "bar": 1`, true},
	}
	d := cloudformationDetector(0)
	for _, tc := range testCases {
		m := &Manifest{}
		content := &Content{
			path: tc.name,
			Head: []byte(tc.content),
		}
		d.DetectContent(m, tc.name, content)
		if tc.match && (m.CloudformationFiles.Len() != 1 || m.CloudformationFiles.Get(0) != tc.name) {
			t.Error(tc, m.CloudformationFiles)
		} else if !tc.match && m.CloudformationFiles.Len() != 0 {
			t.Error(tc, m.CloudformationFiles)
		}
	}
}
