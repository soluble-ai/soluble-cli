package trivyfs

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
	assert.Equal("GHSA-g95f-p29q-9xw4", findings[0].Tool["VulnerabilityID"])
}
