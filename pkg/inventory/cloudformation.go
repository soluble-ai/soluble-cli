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
	"strings"
)

type cloudformationDetector int

var _ FileDetector = cloudformationDetector(0)

func (cloudformationDetector) DetectFileName(m *Manifest, path string) ContentDetector {
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".json") {
		return cloudformationDetector(0)
	}
	return nil
}

func (cloudformationDetector) DetectContent(m *Manifest, path string, buf []byte) {
	d := Decode(path, buf)
	if _, ok := d["AWSTemplateFormatVersion"]; ok {
		m.CloudformationFiles.Add(path)
	}
}
