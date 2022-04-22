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

package secrets

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestParseResults(t *testing.T) {
	assert := assert.New(t)
	results, err := util.ReadJSONFile("testdata/results.json")
	assert.Nil(err)
	tool := &Tool{}
	assert.Nil(tool.Validate())
	result := tool.parseResults(&tools.Result{}, results)
	assert.Equal(2, len(result.Findings))
	f := findFinding(result.Findings, "go.sum")
	assert.NotNil(f)
	assert.Equal("go.sum", f.FilePath)
	assert.Equal("Base64 High Entropy String", f.Title)
	assert.Equal(2, f.Line)
	assert.Equal(results.Unwrap(), result.Data.Unwrap())
	tool.Exclude = []string{"go.sum"}
	assert.Nil(tool.Validate())
	result = tool.parseResults(&tools.Result{}, results)
	assert.Equal(1, len(result.Findings))
	assert.Equal(1, result.Data.Path("results").Size())
}

func findFinding(findings assessments.Findings, filename string) *assessments.Finding {
	for _, f := range findings {
		if f.FilePath == filename {
			return f
		}
	}
	return nil
}
