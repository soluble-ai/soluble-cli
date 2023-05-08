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

func TestSingleAssessmentHelmScan(t *testing.T) {
	assert := assert.New(t)
	tool := test.NewTool(t, "helm-scan", "--include", "**", "--parallel", "-2", "-d", "../../../pkg/tools/checkov/testdata/helm-charts",
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

func TestParallelScanResults(t *testing.T) {
	// run parallel scan and compare results against equivalent set of sequential scans
	assert := assert.New(t)
	// parallel scan
	tool := test.NewTool(t, "helm-scan", "--include", "**", "--parallel", "-2", "-d", "../../../pkg/tools/checkov/testdata/helm-charts",
		"--format", "json").WithFingerprints()
	tool.Must(tool.Run())
	n := tool.JSON()
	assert.Equal(1, n.Size())
	parallelFindings := n.Get(0).Path("findings")

	// sequential scan 1 scan
	tool = test.NewTool(t, "helm-scan", "-d", "../../../pkg/tools/checkov/testdata/helm-charts/charts/subchart",
		"--format", "json").WithFingerprints()
	tool.Must(tool.Run())
	n = tool.JSON()
	assert.Equal(1, n.Size())
	seqScan1Findings := n.Get(0).Path("findings")

	// sequential scan 1 scan
	tool = test.NewTool(t, "helm-scan", "-d", "../../../pkg/tools/checkov/testdata/helm-charts/charts/subchart2",
		"--format", "json").WithFingerprints()
	tool.Must(tool.Run())
	n = tool.JSON()
	assert.Equal(1, n.Size())
	seqScan2Findings := n.Get(0).Path("findings")

	// parallel results should match equivalent set of sequential scans
	assert.Equal(parallelFindings.Size(), seqScan1Findings.Size()+seqScan2Findings.Size())
}
