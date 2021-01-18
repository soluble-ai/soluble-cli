package cfnnag

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestParseResults(t *testing.T) {
	assert := assert.New(t)
	results, err := util.ReadJSONFile("testdata/results.json")
	assert.Nil(err)
	findings := parseResults(results)
	assert.Equal(2, len(findings))
	assert.Equal("F1000", findings[0].Tool["id"])
	assert.Equal("FAIL", findings[0].Tool["type"])
	assert.Equal(49, findings[0].Line)
}
