package secrets

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
	result := tool.parseResults(results)
	assert.Equal(3, len(result.Findings))
	assert.Equal("Base64 High Entropy String", result.Findings[0].Title)
	assert.NotEqual(0, result.Findings[0].Line)
	assert.Equal("dockerfiles/node-docker-demo/package-lock.json", result.Findings[0].FilePath)
}
