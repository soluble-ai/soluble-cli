//go:build integration

package integration

import (
	"path/filepath"
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
	"github.com/soluble-ai/soluble-cli/pkg/repotree"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestHelmScan(t *testing.T) {
	assert := assert.New(t)
	tool := test.NewTool(t, "helm-scan", "-d", "../../../pkg/tools/checkov/testdata/mychart",
		"--format", "json").WithFingerprints()
	tool.Must(tool.Run())
	n := tool.JSON()
	assert.Equal(1, n.Size())
	findings := n.Get(0).Path("findings")
	assert.Greater(findings.Size(), 30)
	repoRoot, err := repotree.FindRepoRoot("")
	assert.NoError(err)
	for _, fg := range tool.Fingerprints.Elements() {
		file := filepath.Join(repoRoot, fg.Path("repoPath").AsText())
		assert.True(util.FileExists(file), "%s should exist", file)
	}
}
