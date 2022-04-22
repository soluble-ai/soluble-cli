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

package tfsec

import (
	"fmt"
	"os"
	"testing"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/stretchr/testify/assert"
)

func TestParseResults(t *testing.T) {
	assert := assert.New(t)
	output, err := os.ReadFile("testdata/results.json.txt")
	assert.NoError(err)
	output = trimOutput(output)
	results, err := jnode.FromJSON(output)
	assert.Nil(err)
	fmt.Println(results)
	tool := &Tool{}
	tool.Directory = "/x/work/solublegoat/terraform/aws"
	tool.RepoRoot = "/x/work/solublegoat"
	result := &tools.Result{}
	tool.parseResults(result, results)
	assert.Equal(9, len(result.Findings))
	f := result.Findings[8]
	assert.Equal(16, f.Line)
	assert.Equal("variables.tf", f.FilePath)
	// verify filepath was rewritten within results.Data
	assert.Equal("variables.tf", result.Data.Path("results").Get(8).Path("location").Path("filename").AsText())
}
