package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/policy/policyimporter"
	"github.com/soluble-ai/soluble-cli/pkg/policy/testutil"
	"github.com/stretchr/testify/assert"
)

func cleanUp() {
	os.RemoveAll("testdata/tmp")
}
func Test_converter(t *testing.T) {
	assert := assert.New(t)
	defer cleanUp()

	input := "testdata/input/policies"
	dest := "testdata/tmp/policies/opal"
	regoFiles := policyimporter.Find(input, ".rego")

	for i := len(regoFiles) - 1; i >= 0; i-- {
		fmt.Println("rego path: ", regoFiles[i])
		p := policyimporter.Policy{Tool: "opal"}
		if err := p.Convert(regoFiles[i], dest); err != nil {
			t.Fail()
		}
	}

	// test structure
	lwStructure := policyimporter.Find(dest, ".rego")
	assert.Equal(lwStructure[0], "testdata/tmp/policies/opal/s3_https_access/cloudformation/policy.rego")
	assert.Equal(lwStructure[1], "testdata/tmp/policies/opal/s3_https_access/terraform/policy.rego")

	// test metadata
	expectedFilePath := "testdata/expected/metadata.yaml"
	actualFilePath := "testdata/tmp/policies/opal/s3_https_access/metadata.yaml"
	diff := testutil.CompareYamlFiles(actualFilePath, expectedFilePath)
	assert.Equal(0, diff)
}
