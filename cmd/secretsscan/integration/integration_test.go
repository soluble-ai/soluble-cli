//go:build integration

package integration

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
)

func TestSecretsScan(t *testing.T) {
	test.RequireAPIToken(t)
	tool := test.NewTool(t, "secrets-scan", "--exclude", "go.sum", "--exclude", "pkg/**/testdata/*.json",
		"--exclude", "pkg/tools/cloudsploit/**", "--error-not-empty").WithUpload(true).WithRepoRootDir()
	tool.Must(tool.Run())
}
