package secrets

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestParseResults(t *testing.T) {
	assert := assert.New(t)
	results, err := util.ReadJSONFile("testdata/results.json")
	assert.Nil(err)
	tool := &Tool{}
	assert.Nil(tool.Validate())
	result := tool.parseResults(results)
	assert.Equal(2, len(result.Findings))
	f := findFinding(result.Findings, "go.sum")
	assert.NotNil(f)
	assert.Equal("go.sum", f.FilePath)
	assert.Equal("Base64 High Entropy String", f.Title)
	assert.Equal(2, f.Line)
	assert.Equal(results.Unwrap(), result.Data.Unwrap())
	tool.Exclude = []string{"go.sum"}
	assert.Nil(tool.Validate())
	result = tool.parseResults(results)
	assert.Equal(1, len(result.Findings))
	assert.Equal(1, result.Data.Path("results").Size())
}

func findFinding(findings assessments.Findings, filename string) *assessments.Finding {
	for _, f := range findings {
		if f.FilePath == filename {
			return f
		}
	}
	return nil
}
