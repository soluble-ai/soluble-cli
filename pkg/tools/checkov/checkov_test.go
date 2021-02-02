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
