package test

import (
	"os"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/policy/testutil"

	"github.com/soluble-ai/soluble-cli/pkg/policy/custompolicybuilder"

	"github.com/stretchr/testify/assert"
)

func setupDirPath(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	return err
}

// Clean up after test
func cleanUpPoliciesDir() {
	os.RemoveAll("policies")
}

func TestCreate_ExpectedMetadataYaml(t *testing.T) {
	pt := custompolicybuilder.PolicyTemplate{
		Name:      "unit_test_cust_policy",
		CheckType: "terraform",
		Dir:       "policies",
		Tool:      "opal",
		Desc:      "unit test policy",
		Severity:  "Info",
		Title:     "unit test custom policy",
		Provider:  "AWS",
		Category:  "General",
	}
	assert := assert.New(t)
	if err := setupDirPath("policies"); err != nil {
		t.Fail()
	}
	defer cleanUpPoliciesDir()

	err := pt.CreateCustomPolicyTemplate()
	assert.NoError(err)

	// Compare Metadata.yaml
	actualFilePath := "policies/opal/unit_test_cust_policy/metadata.yaml"
	expectedFilePath := "testdata/policies/opal/unit_test_cust_policy/metadata.yaml"

	diff := testutil.CompareYamlFiles(actualFilePath, expectedFilePath)
	assert.Equal(0, diff)
}
