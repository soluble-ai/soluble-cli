package test

import (
	"github.com/soluble-ai/soluble-cli/pkg/policy/policyimporter"
	"github.com/soluble-ai/soluble-cli/pkg/policy/testutil"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func cleanUp() {
	os.RemoveAll("testdata/tmp")
}


func exists(file string) error {
	_, err := os.Stat("testdata/tmp/policies/opal/" + file)
	return err
}

func Test_converter(t *testing.T) {
	assert := assert.New(t)
	cleanUp()
	defer cleanUp()

	converter := &policyimporter.Converter{
		OpalRegoPath :  "testdata/input/policies",
		DestPath: "testdata/tmp/policies/opal",
		TestPath: "testdata/input/policiesTest/tests/policies",
	}
	if err := converter.ConvertOpalBuiltIns(); err != nil {
		t.Fail()
	}

	//test metadata
	expectedFilePath := "testdata/expected/metadata.yaml"
	actualFilePath := "testdata/tmp/policies/opal/s3_https_access/metadata.yaml"
	diff := testutil.CompareYamlFiles(actualFilePath, expectedFilePath)
	assert.Equal(0, diff)

	// test dir structure
	// test all policies were created
	assert.NoError(exists("s3_block_public_access/cloudformation/policy.rego"))
	assert.NoError(exists("s3_encryption/cloudformation/policy.rego"))

	// check test policies exist
	assert.NoError(exists("s3_block_public_access/cloudformation/tests/policy_test.rego"))
	assert.NoError(exists("s3_encryption/cloudformation/tests/policy_test.rego"))

	// check relevant input files exist
	assert.NoError(exists("s3_block_public_access/cloudformation/tests/inputs/invalid_block_public_access_infra.yaml"))
	assert.NoError(exists("s3_block_public_access/cloudformation/tests/inputs/invalid_block_public_access_infra_yaml.rego"))
	assert.NoError(exists("s3_block_public_access/cloudformation/tests/inputs/valid_block_public_access_infra.yaml"))
	assert.NoError(exists("s3_block_public_access/cloudformation/tests/inputs/valid_block_public_access_infra_yaml.rego"))
}

