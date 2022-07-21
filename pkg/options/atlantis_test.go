package options

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestAtlantis(t *testing.T) {
	assert := assert.New(t)
	p := &PrintOpts{
		OutputFormat: "atlantis",
	}
	n, err := util.ReadJSONFile("testdata/assessment.json.gz")
	assert.NoError(err)
	s := &strings.Builder{}
	p.outputSource = func() io.Writer {
		return s
	}
	p.PrintResult(n)
	exp, _ := os.ReadFile("testdata/atlantis.txt")
	assert.Equal(string(exp), s.String())
}
