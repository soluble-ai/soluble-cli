//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
	"github.com/stretchr/testify/assert"
)

func TestMultiDocument(t *testing.T) {
	assert := assert.New(t)
	cmd := test.NewTool(t, "k8s-scan", "--config-file", "/dev/null", "-d", "testdata/k", "--format", "json").
		WithFingerprints()
	cmd.Must(cmd.Run())
	assert.Equal(2, cmd.Fingerprints.Size())
	for _, n := range cmd.Fingerprints.Elements() {
		assert.True(strings.HasSuffix(n.Path("filePath").AsText(), "k8s.yaml"), n)
		repoPath := n.Path("repoPath").AsText()
		assert.Equal("cmd/k8sscan/integration/testdata/k/k8s.yaml", repoPath)
		assert.True(n.Path("multiDocumentFile").AsBool(), "should be multi-document - %s", n)
		assert.NotEmpty(n.Path("partialFingerprint").AsText(), "fingeprint should not be empty - %s", n)
	}
}
