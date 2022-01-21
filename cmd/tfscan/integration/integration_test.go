//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
	"github.com/stretchr/testify/assert"
)

func TestScan(t *testing.T) {
	tool := test.NewTool(t, "tf-scan", "-d", "testdata", "--config-file", "/dev/null")
	tool.Must(tool.Run())
	lines := strings.Split(tool.Out.String(), "\n")
	assert := assert.New(t)
	assert.Greater(len(lines), 1)
	assert.Contains(lines[0], "SID ")
}

func TestScanUploadJSON(t *testing.T) {
	test.RequireAPIToken(t)
	tool := test.NewTool(t, "tf-scan", "-d", "testdata", "--config-file", "/dev/null", "--format", "json").
		WithUpload(true)
	tool.Must(tool.Run())
	n := tool.JSON()
	assert := assert.New(t)
	if !assert.Equal(1, n.Size()) {
		return
	}
	assmt := n.Get(0)
	assert.NotEmpty(assmt.Path("appUrl").AsText())
	assert.NotEmpty(assmt.Path("assessmentId").AsText())
	assert.Greater(assmt.Path("findings").Size(), 1)
}
