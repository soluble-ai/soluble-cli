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
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

type terraformDetector struct {
	dotTerraform string
}

var providerRegexp = regexp.MustCompile(`(?m)^provider\s+"`)

var _ FileDetector = &terraformDetector{}

func (d *terraformDetector) DetectFileName(m *Manifest, path string) ContentDetector {
	if d.dotTerraform != "" {
		if strings.HasPrefix(path, d.dotTerraform) {
			return nil
		}
		d.dotTerraform = ""
	}
	if strings.HasSuffix(path, ".tf") || strings.HasSuffix(path, ".tf.json") {
		m.TerraformModules.Add(filepath.Dir(path))
		return d
	}
	return nil
}

func (d *terraformDetector) DetectDirName(m *Manifest, path string) {
	// terraform init will download modules into .terraform, we're going to
	// ignore these
	if d.dotTerraform == "" && filepath.Base(path) == ".terraform" {
		d.dotTerraform = path
	}
}

func (*terraformDetector) DetectContent(m *Manifest, path string, content []byte) {
	if strings.HasSuffix(path, ".tf") {
		if providerRegexp.Find(content) != nil {
			m.TerraformRootModules.Add(filepath.Dir(path))
		}
	} else {
		p := gjson.ParseBytes(content).Get("provider")
		if p.IsArray() || p.IsObject() {
			m.TerraformRootModules.Add(filepath.Dir(path))
		}
	}
}
