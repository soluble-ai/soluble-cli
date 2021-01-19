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
	result := tool.parseResults(results)
	assert.Equal(4, len(result.Findings))
	f := result.Findings[0]
	assert.Equal("infrastructure.tf", f.FilePath)
	assert.Equal(1, f.Line)
	assert.Equal("MEDIUM", f.Tool["severity"])
}
