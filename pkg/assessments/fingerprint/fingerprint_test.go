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

package fingerprint

import (
	"bufio"
	"os"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestPartialFingerprints(t *testing.T) {
	assert := assert.New(t)
	f, err := os.Open("fingerprint.go")
	util.Must(err)
	defer f.Close()
	r := bufio.NewReader(f)
	fingerprints := map[int]string{}
	err = Partial(r, func(n int, fingerprint string) {
		fingerprints[n] = fingerprint
	})
	assert.Nil(err)
	// check that we got a fingerprint for every line
	assert.Greater(len(fingerprints), 50)
	for i := 1; i <= len(fingerprints); i++ {
		fingerprint := fingerprints[i]
		assert.NotEmpty(fingerprint)
		assert.Greater(len(fingerprint), 15)
	}
}
