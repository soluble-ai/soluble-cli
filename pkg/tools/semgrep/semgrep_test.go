package semgrep

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestParseResults(t *testing.T) {
	assert := assert.New(t)
	n, err := util.ReadJSONFile("testdata/results.json")
	assert.Nil(err)
	tool := &Tool{}
	assert.Nil(tool.Validate())
	result := tool.parseResults(n)
	assert.Equal(2, len(result.Findings))
	f := result.Findings[0]
	assert.Equal(46, f.Line)
	assert.Equal("pdl/src/main/java/pdl/PdlDiag.java", f.FilePath)
	assert.Equal("-", f.Tool["check_id"])
	assert.Equal(n.Unwrap(), result.Data.Unwrap())
}
