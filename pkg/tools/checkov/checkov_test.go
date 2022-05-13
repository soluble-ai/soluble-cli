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
	"path/filepath"
	"strings"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestParseResults(t *testing.T) {
	assert := assert.New(t)
	tool := &Tool{}
	assert.NoError(tool.Validate())
	results, err := util.ReadJSONFile("testdata/results.json.gz")
	assert.NoError(err)
	result := tool.processResults(&tools.Result{}, results)
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

func TestParseResults2(t *testing.T) {
	assert := assert.New(t)
	tool := &Tool{}
	assert.NoError(tool.Validate())
	results, err := util.ReadJSONFile("testdata/results2.json.gz")
	assert.NoError(err)
	result := tool.processResults(&tools.Result{}, results)
	assert.Equal("6", result.Values["RESOURCE_COUNT"])
}

func TestPropagateTfvarsEnv(t *testing.T) {
	d := &tools.DockerTool{}
	propagateTfVarsEnv(d, []string{"TF_VAR_foo=foo", "TF_VAR_bar=bar", "PATH=/bin;/usr/bin"})
	assert.ElementsMatch(t, []string{"TF_VAR_foo", "TF_VAR_bar"}, d.PropagateEnvironmentVars)
}

func TestParseHelmResults(t *testing.T) {
	assert := assert.New(t)
	for _, resultsFile := range []string{
		"testdata/helm-results-local.json.gz",
		"testdata/helm-results-docker.json.gz",
	} {
		tool := &Tool{Framework: "helm"}
		tool.Directory = "testdata/mychart"
		assert.NoError(tool.Validate())
		results, err := util.ReadJSONFile(resultsFile)
		assert.NoError(err)
		result := tool.processResults(&tools.Result{}, results)
		assert.Greater(len(result.Findings), 30)
		for _, finding := range result.Findings {
			assert.False(strings.HasPrefix(finding.FilePath, "/"), "%s is relative", finding.FilePath)
			assert.False(strings.HasPrefix(finding.FilePath, "tmp/"), "%s is relative", finding.FilePath)
			file := filepath.Join(tool.Directory, finding.FilePath)
			assert.True(util.FileExists(file), "%s : %s should exist", resultsFile, file)
		}
	}
}
