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
	// As of 5-24-22 the gcr.io/soluble-repo version of checkov doesn't
	// support kustomize, so run with a specific later version that does.
	// This will need to get sorted out before kustomize-scan is GA.
	tool := test.NewTool(t, "ea", "kustomize-scan", "-d", "../../k8sscan/integration/testdata/kust",
		"--tool-version", "bridgecrew/checkov:2.0.1140", "--use-empty-config-file").WithFingerprints()
	tool.Must(tool.Run())
	repoRoot, err := repotree.FindRepoRoot("")
	assert.NoError(err)
	for _, fg := range tool.Fingerprints.Elements() {
		file := filepath.Join(repoRoot, fg.Path("repoPath").AsText())
		assert.True(util.FileExists(file), "%s should exist", file)
	}
}
