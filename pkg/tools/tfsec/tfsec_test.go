package tfsec

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestParseResults(t *testing.T) {
	assert := assert.New(t)
	results, err := util.ReadJSONFile("testdata/results.json")
	assert.Nil(err)
	tool := &Tool{}
	tool.Directory = "/Users/samshen/work/solublegoat/terraform/aws"
	result := tool.parseResults(results)
	assert.Equal(9, len(result.Findings))
	f := result.Findings[8]
	assert.Equal(16, f.Line)
	assert.Equal("variables.tf", f.FilePath)
	// verify filepath was rewritten within results.Data
	assert.Equal("variables.tf", result.Data.Path("results").Get(8).Path("location").Path("filename").AsText())
}
