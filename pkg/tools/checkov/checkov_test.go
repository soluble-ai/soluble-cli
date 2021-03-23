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

package checkov

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestParseResults(t *testing.T) {
	assert := assert.New(t)
	tool := &Tool{}
	assert.Nil(tool.Validate())
	results, err := util.ReadJSONFile("testdata/results.json")
	assert.Nil(err)
	result := tool.processResults(results)
	assert.Equal(7, len(result.Findings))
	passed := 0
	for _, f := range result.Findings {
		if f.Pass {
			passed++
			if passed == 1 {
				assert.Equal("CKV_AWS_24", f.Tool["check_id"])
				assert.Equal("security.tf", f.FilePath)
				assert.Equal(1, f.Line)
			}
		}
	}
	assert.Equal(6, passed)
	assert.Equal(results.Unwrap(), result.Data.Unwrap())
}
