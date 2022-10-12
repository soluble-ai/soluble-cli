package print

import (
	"bytes"
	"testing"

	"github.com/soluble-ai/go-jnode"
	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	assert := assert.New(t)
	w := &bytes.Buffer{}
	n := jnode.NewArrayNode().Append("test").Append("hemanth")
	printer, err := NewTemplatePrinter("Hello {{len .}}!.")
	if assert.NoError(err) {
		printer.PrintResult(w, n)
		assert.Equal("Hello 2!.", w.String())
	}
	printer, err = NewTemplatePrinter("@testdata/len.tmpl")
	if assert.NoError(err) {
		w.Reset()
		printer.PrintResult(w, n)
		assert.Equal("2", w.String())
	}
}
