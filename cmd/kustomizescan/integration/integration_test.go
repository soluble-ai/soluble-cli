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

func TestKustomizeScan(t *testing.T) {
	assert := assert.New(t)
	tool := test.NewTool(t, "kustomize-scan", "-d", "../../k8sscan/integration/testdata/kust",
		"--use-empty-config-file").WithFingerprints()
	tool.Must(tool.Run())
	repoRoot, err := repotree.FindRepoRoot("")
	assert.NoError(err)
	for _, fg := range tool.Fingerprints.Elements() {
		file := filepath.Join(repoRoot, fg.Path("repoPath").AsText())
		assert.True(util.FileExists(file), "%s should exist", file)
	}
}

func TestKustomizeUpload(t *testing.T) {
	assert := assert.New(t)
	tool := test.NewTool(t, "kustomize-scan", "-d", "../../k8sscan/integration/testdata/kust",
		"--use-empty-config-file", "--format", "json").WithUpload(true)
	tool.Must(tool.Run())
	n := tool.JSON()
	findings := n.Get(0).Path("findings")
	assert.Greater(findings.Size(), 30)
}
