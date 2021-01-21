package cfnpythonlint

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestParseResults(t *testing.T) {
	assert := assert.New(t)
	results, err := util.ReadJSONFile("testdata/results.json")
	assert.Nil(err)
	result := parseResults(results)
	assert.Equal(1, len(result.Findings))
	f := result.Findings[0]
	assert.Equal("EC2InstanceWithSecurityGroupSample.yaml", f.FilePath)
	assert.Equal(25, f.Line)
	assert.LessOrEqual(100, len(f.Tool["Message"]))
	assert.Equal(results.Unwrap(), result.Data.Unwrap())
}
