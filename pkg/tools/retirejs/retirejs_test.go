package retirejs

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestParseResults(t *testing.T) {
	assert := assert.New(t)
	results, err := util.ReadJSONFile("testdata/results.json.gz")
	assert.NoError(err)
	assert.NotNil(results)
	tool := &Tool{}
	tool.Directory = "/Users/samshen/3rd/retire.js"
	tool.NoDocker = true
	tool.Exclude = []string{"**/retire-example*.js"}
	assert.NoError(tool.Validate())
	r := tool.parseResults(results)
	assert.NotNil(r)
	assert.Equal(9, len(r.Findings))
	f := r.Findings[0]
	assert.Equal("medium", f.Tool["severity"])
	assert.Equal("PR-307", f.Tool["identifier"])
	assert.Equal(2, r.Data.Path("data").Size())
}
