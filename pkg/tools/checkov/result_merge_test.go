package checkov

import (
	"testing"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestCombineArrays(t *testing.T) {
	assert := assert.New(t)
	r1 := util.MustReadJSONFile("testdata/results.json.gz")
	r2 := util.MustReadJSONFile("testdata/results2.json.gz")
	assert.Equal(6, r1.Path("results").Path("passed_checks").Size())
	assert.Equal(5, r2.Path("results").Path("passed_checks").Size())
	combineArrays(r1, r2, "results", "passed_checks")
	assert.Equal(11, r1.Path("results").Path("passed_checks").Size())
	n := jnode.NewObjectNode()
	assert.Equal(1, r1.Path("results").Path("failed_checks").Size())
	combineArrays(n, r1, "results", "failed_checks")
	assert.Equal(1, r1.Path("results").Path("failed_checks").Size())
	combineArrays(n, r1, "foo")
}

func TestMergeResults(t *testing.T) {
	assert := assert.New(t)
	r1 := util.MustReadJSONFile("testdata/results.json.gz")
	r2 := util.MustReadJSONFile("testdata/results2.json.gz")
	mergeResults(r1, r2)
	assert.Equal(17, r1.Path("summary").Path("passed").AsInt())
}
