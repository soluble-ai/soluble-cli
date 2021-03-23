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

package terrascan

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestParseResults(t *testing.T) {
	assert := assert.New(t)
	results, err := util.ReadJSONFile("testdata/results.json")
	assert.Nil(err)
	tool := &Tool{
		DirectoryBasedToolOpts: tools.DirectoryBasedToolOpts{
			Exclude: []string{"nat-server.tf"},
		},
	}
	assert.Nil(tool.Validate())
	result := tool.parseResults(results)
	assert.Equal(4, len(result.Findings))
	f := result.Findings[0]
	assert.Equal("infrastructure.tf", f.FilePath)
	assert.Equal(1, f.Line)
	assert.Equal("MEDIUM", f.Tool["severity"])
	assert.Equal(results.Unwrap(), result.Data.Unwrap())
}
