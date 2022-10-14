package test

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/soluble-ai/soluble-cli/cmd/policy"
	"github.com/stretchr/testify/assert"
)

// helper funcs
func executeCreate(args []string) error {
	cmd := policy.CreateCommand()
	cmd.SetArgs(args)
	err := cmd.Execute()
	return err
}

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

// User input validation
func TestCreate_NoCurrentPoliciesDir(t *testing.T) {
	assert := assert.New(t)
	err := executeCreate([]string{"--name", "unit_test_policy",
		"--check-type", "terraform",
		"--type", "opal"})
	assert.EqualError(err, "could not find 'policies' directory in current directory."+
		"\ncreate 'policies' directory or use -d to target an existing policies directory")
}

func TestCreate_InvalidDirPath(t *testing.T) {
	assert := assert.New(t)
	err := executeCreate([]string{"--name", "unit_test_policy",
		"--check-type", "terraform",
		"--type", "opal",
		"-d", "./badpath"})
	assert.EqualError(err, "invalid directory path: ./badpath\nprovide path to existing policies directory")
}

func TestCreate_DirPathNotFound(t *testing.T) {
	assert := assert.New(t)
	err := executeCreate([]string{"--name", "unit_test_policy",
		"--check-type", "terraform",
		"--type", "opal",
		"-d", "./policies"})
	assert.EqualError(err, "could not find directory: ./policies\ntarget an existing policies directory.")
}

func TestCreate_InvalidPolicyName(t *testing.T) {
	assert := assert.New(t)
	err := executeCreate([]string{"--name", "bad.policy-name",
		"--check-type", "terraform",
		"--type", "opal"})
	assert.EqualError(err, "invalid name: bad.policy-name. name must consist only of [a-z0-9-]")
}

func TestCreate_InvalidPolicyType(t *testing.T) {
	assert := assert.New(t)
	err := executeCreate([]string{"--name", "unit_test_policy",
		"--check-type", "terraform",
		"--type", "nope"})
	assert.Error(err)
	errString := strings.Split(err.Error(), ":")
	assert.Equal("invalid type. type is one of", errString[0])
	assert.Contains(errString[1], "checkov-py")
	assert.Contains(errString[1], "opal")
	assert.Contains(errString[1], "checkov")
}

func TestCreate_InvalidPolicyCheckType(t *testing.T) {
	assert := assert.New(t)
	err := executeCreate([]string{"--name", "unit_test_policy",
		"--check-type", "nope",
		"--type", "opal"})
	assert.EqualError(err, "invalid check-type. check-type is one of: [terraform terraform-plan cloudformation kubernetes helm docker secrets]")
}

func TestCreate_InvalidPolicySeverity(t *testing.T) {
	assert := assert.New(t)
	err := executeCreate([]string{"--name", "unit_test_policy",
		"--check-type", "terraform",
		"--type", "opal",
		"--severity", "nope"})
	assert.EqualError(err, "invalid severity 'nope'. severity is one of: [info low medium high critical]")
}

func TestCreate_CustPolicyExists(t *testing.T) {
	assert := assert.New(t)
	//create dummy cust policy dir
	if err := setupDirPath("policies/opal/unit_test_policy/terraform"); err != nil {
		t.Fail()
	}
	err := executeCreate([]string{"--name", "unit_test_policy",
		"--check-type", "terraform",
		"--type", "opal"})
	assert.EqualError(err, "custom policy 'unit_test_policy' with check type 'terraform' already exists in directory 'policies/opal/unit_test_policy/terraform'")
	cleanUpPoliciesDir()
}

func TestCreate_ExpectedMetadataYaml(t *testing.T) {
	assert := assert.New(t)
	//create dummy cust policy dir
	if err := setupDirPath("policies"); err != nil {
		t.Fail()
	}
	err := executeCreate([]string{"--name", "unit_test_cust_policy",
		"--check-type", "terraform",
		"--type", "opal",
		"--description", "unit test policy",
		"--severity", "info",
		"--resource-type", "aws_s3_bucket",
		"--title", "unit test custom policy"})
	assert.NoError(err)

	// Compare Metadata.yaml
	actualFilePath := "policies/opal/unit_test_cust_policy/metadata.yaml"
	expectedFilePath := "testdata/policies/opal/unit_test_cust_policy/metadata.yaml"
	actual := readMetadataFile(actualFilePath)
	expected := readMetadataFile(expectedFilePath)

	actData, err := yaml.Marshal(actual)
	expData, err := yaml.Marshal(expected)

	diff := bytes.Compare(actData, expData)
	assert.Equal(0, diff)

	cleanUpPoliciesDir()
}
