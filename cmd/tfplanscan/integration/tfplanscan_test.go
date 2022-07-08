//go:build integration

package integration

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
	"github.com/stretchr/testify/assert"
)

func TestScan(t *testing.T) {
	tool := test.NewTool(t, "ea", "terraform-plan-scan",
		"--plan", "testdata/terraform.tfplan.json", "--format", "json")
	tool.Must(tool.Run())
	n := tool.JSON()
	assert := assert.New(t)
	assert.Equal(1, n.Size())
	findings := n.Get(0).Path("findings")
	assert.Greater(findings.Size(), 1)
}
