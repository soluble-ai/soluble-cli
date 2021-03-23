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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCI(t *testing.T) {
	assert := assert.New(t)
	m := &Manifest{}
	m.scan(filepath.Join("testdata", "ci"), cidetector(0))
	assert.ElementsMatch(m.CISystems.Values(), []string{
		"github", "drone", "gitlab", "circleci", "jenkins", "travis", "azure",
	})
	m.CISystems.Reset()
	m.scan("testdata", cidetector(0))
	if m.CISystems.Len() != 0 {
		t.Error(m.CISystems.Values())
	}
}
