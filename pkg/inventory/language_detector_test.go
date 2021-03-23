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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLanguageDetectors(t *testing.T) {
	assert := assert.New(t)
	m := &Manifest{}
	m.scan("testdata/lang", pythonDetector())
	assert.Contains(m.PythonDirectories.Values(), "python-app")
	m = &Manifest{}
	m.scan("testdata/lang/python-app", pythonDetector())
	assert.Contains(m.PythonDirectories.Values(), ".")
	m = &Manifest{}
	m.scan("testdata/lang", javaAntMavenDetector(), javaGradleDetector())
	assert.ElementsMatch(m.JavaDirectories.Values(), []string{
		"java", "java2", "java3",
	})
	m = &Manifest{}
	m.scan("testdata/lang/java3", javaGradleDetector())
	assert.ElementsMatch(m.JavaDirectories.Values(), []string{"."})
}
