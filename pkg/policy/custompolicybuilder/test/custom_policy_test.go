package test

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/policy"
	"github.com/stretchr/testify/assert"
)

func TestCreateDirectoryStructure_NoCurrentPoliciesDir(t *testing.T) {
	assert := assert.New(t)
	cmd := policy.CreateCommand()
	cmd.SetArgs([]string{"--name", "unit-test-policy", "--check-type", "terraform", "--type", "opal"})
	err := cmd.Execute()
	assert.EqualError(err, "could not find 'policies' directory in current directory."+
		"\ncreate 'policies' directory or use -d to target existing policies directory")
}

// TODO more tests
