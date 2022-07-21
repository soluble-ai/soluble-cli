package opal

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestParseResults(t *testing.T) {
	assert := assert.New(t)
	n := util.MustReadJSONFile("testdata/results.json.gz")
	assert.NotNil(n)
	tool := &Tool{}
	result := &tools.Result{}
	tool.parseResults(result, n)
	assert.Len(result.Findings, 12)
	for _, f := range result.Findings {
		assert.Equal("pkg/tools/opal/testdata/s3/main.tf", f.FilePath)
		assert.NotEmpty(f.Severity)
	}
}
