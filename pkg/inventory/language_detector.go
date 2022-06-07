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

	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type LanguageDetector struct {
	getValues          func(*Manifest) *util.StringSet
	markerFiles        []string
	content            ContentDetector
	collapseNestedDirs bool
}

var _ FileDetector = &LanguageDetector{}
var _ FinalizeDetector = &LanguageDetector{}

func (d *LanguageDetector) DetectFileName(m *Manifest, path string) ContentDetector {
	base := filepath.Base(path)
	for _, mf := range d.markerFiles {
		if mf == base {
			d.getValues(m).Add(filepath.Dir(path))
			return d.content
		}
	}
	return nil
}

func (d *LanguageDetector) FinalizeDetection(m *Manifest) {
	if !d.collapseNestedDirs {
		return
	}
	collapseNestedDirs(d.getValues(m))
}

func pythonDetector() *LanguageDetector {
	return &LanguageDetector{
		getValues:   func(m *Manifest) *util.StringSet { return &m.PythonDirectories },
		markerFiles: []string{"Pipfile", "requirements.txt"},
	}
}

func goDetector() *LanguageDetector {
	return &LanguageDetector{
		getValues:   func(m *Manifest) *util.StringSet { return &m.GODirectories },
		markerFiles: []string{"go.mod"},
	}
}

func javaAntMavenDetector() *LanguageDetector {
	return &LanguageDetector{
		getValues:          func(m *Manifest) *util.StringSet { return &m.JavaDirectories },
		markerFiles:        []string{"pom.xml", "build.xml", "build.gradle"},
		collapseNestedDirs: true,
	}
}

func javaGradleDetector() *LanguageDetector {
	return &LanguageDetector{
		getValues:   func(m *Manifest) *util.StringSet { return &m.JavaDirectories },
		markerFiles: []string{"build.gradle"},
	}
}

func rubyDetector() *LanguageDetector {
	return &LanguageDetector{
		getValues:   func(m *Manifest) *util.StringSet { return &m.RubyDirectories },
		markerFiles: []string{"Gemfile"},
	}
}

func nodeDetector() *LanguageDetector {
	return &LanguageDetector{
		getValues:   func(m *Manifest) *util.StringSet { return &m.NodeDirectories },
		markerFiles: []string{"package-lock.json", "yarn.lock"},
	}
}
