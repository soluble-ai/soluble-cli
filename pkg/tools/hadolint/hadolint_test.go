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

package hadolint

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestParseResults(t *testing.T) {
	assert := assert.New(t)
	results, err := util.ReadJSONFile("testdata/results.json.gz")
	assert.Nil(err)
	tool := &Tool{}
	result := &tools.Result{}
	tool.parseResults(result, results)
	assert.Equal(2, len(result.Findings))
	assert.Equal(2, result.Data.Size())
	f := result.Findings[0].Tool
	assert.Equal("DL3027", f["rule_id"])
	assert.Equal(results.Unwrap(), result.Data.Unwrap())
}
