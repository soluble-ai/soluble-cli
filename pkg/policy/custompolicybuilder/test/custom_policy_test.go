package test

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/policy/custompolicybuilder"

	"gopkg.in/yaml.v3"

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

func readMetadataFile(path string) map[interface{}]interface{} {
	f, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	data := make(map[interface{}]interface{})

	if err := yaml.Unmarshal(f, &data); err != nil {
		log.Fatal(err)
	}
	return data
}

func TestCreate_ExpectedMetadataYaml(t *testing.T) {
	pt := custompolicybuilder.PolicyTemplate{
		Name:      "unit_test_cust_policy",
		CheckType: "terraform",
		Dir:       "policies",
		Tool:      "opal",
		Desc:      "unit test policy",
		Severity:  "info",
		Title:     "unit test custom policy",
		Provider:  "aws",
		Category:  "general",
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
	actual := readMetadataFile(actualFilePath)
	expected := readMetadataFile(expectedFilePath)

	actData, _ := yaml.Marshal(actual)
	expData, _ := yaml.Marshal(expected)

	diff := bytes.Compare(actData, expData)
	assert.Equal(0, diff)
}
