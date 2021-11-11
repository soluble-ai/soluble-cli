package trivy

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestGetDatav020(t *testing.T) {
	assert := assert.New(t)
	n := util.MustReadJSONFile("testdata/v0.20.2.json.gz")
	data := getData("v0.20.2", n)
	assert.True(data.Path("Vulnerabilities").IsArray())
}

func TestGetData(t *testing.T) {
	assert := assert.New(t)
	n := util.MustReadJSONFile("testdata/v0.18.3.json.gz")
	data := getData("v0.18.3", n)
	assert.True(data.Path("Vulnerabilities").IsArray())
}
