//go:build integration

package integration

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
	"github.com/stretchr/testify/assert"
)

func TestSecretsScan(t *testing.T) {
	test.RequireAPIToken(t)
	tool := test.NewTool(t, "secrets-scan", "--exclude", "go.sum", "--exclude", "pkg/**/testdata/*.json",
		"--exclude", "pkg/tools/cloudsploit/**", "--format", "count").WithUpload(true).WithRepoRootDir()
	tool.Must(tool.Run())
	assert.Equal(t, "0\n", tool.Out.String())
}
