//go:build integration

package integration

import (
	"strconv"
	"strings"
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
	"github.com/stretchr/testify/assert"
)

func TestSecretsScan(t *testing.T) {
	test.RequireAPIToken(t)
	tool := test.NewTool(t, "secrets-scan", "--format", "count").WithUpload(true).WithRepoRootDir()
	tool.Must(tool.Run())
	n, err := strconv.Atoi(strings.TrimSpace(tool.Out.String()))
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, n, 0)
}
