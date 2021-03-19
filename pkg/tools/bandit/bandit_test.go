package bandit

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
	assert.Equal(19, len(result.Findings))
	f := result.Findings[0]
	assert.Equal("./tests/terraform/module_loading/loaders/test_local_path_loader.py", f.FilePath)
	assert.Equal(11, f.Line)
	assert.Equal(results.Unwrap(), result.Data.Unwrap())
}
