//go:build integration

package integration

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
	"github.com/stretchr/testify/assert"
)

func TestPrintTemplate(t *testing.T) {
	assert := assert.New(t)
	cmd := test.NewCommand(t, "print", "--print-template", "@testdata/print.tmpl", "testdata/data.json")
	cmd.Must(cmd.Run())
	assert.Equal("X = 1 Y = 2\nX = 3 Y = 4\n", cmd.Out.String())
}
