//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
	"github.com/stretchr/testify/assert"
)

func TestPolicyVet(t *testing.T) {
	vet := test.NewCommand(t, "policy", "vet",
		"-d", "../../../pkg/policy/checkov/testdata", "--format", "json")
	vet.Must(vet.Run())
	n := vet.JSON()
	assert := assert.New(t)
	assert.Equal(2, n.Path("valid").AsInt(), n)
	assert.Equal(0, n.Path("invalid").AsInt(), n)
}

func TestPolicyTest(t *testing.T) {
	test := test.NewCommand(t, "policy", "test",
		"-d", "../../../pkg/policy/checkov/testdata", "--format", "json")
	test.Must(test.Run())
	n := test.JSON()
	assert := assert.New(t)
	assert.Equal(4, n.Path("passed").AsInt(), n)
	assert.Equal(0, n.Path("failed").AsInt(), n)
}

func TestPolicyConvertHidden(t *testing.T) {
	test := test.NewCommand(t, "policy", "--help")
	test.Must(test.Run())
	assert := assert.New(t)
	assert.False(strings.Contains(test.Out.String(), "convert"))
}

func TestPolicyUploadFail(t *testing.T) {
	assert := assert.New(t)
	test := test.NewCommand(t, "policy", "upload",
		"-d", "path/to/policies")
	err := test.Run()
	assert.Errorf(err,
		"no policies found"+
			"\n\t - Ensure path provided points to the parent directory of the /policies directory"+
			"\n\t - or use --allow-empty to upload no policies")
}
