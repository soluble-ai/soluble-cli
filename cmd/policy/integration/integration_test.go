//go:build integration

package integration

import (
	"github.com/ActiveState/vt10x"
	"github.com/Netflix/go-expect"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

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

func TestPolicyCreateWizard(t *testing.T) {

	assert := assert.New(t)

	// Create a virtual terminal
	c, _, err := vt10x.NewVT10XConsole()
	assert.NoError(err)
	defer c.Close()

	// Define command for `policy create` and attach it to the virtual terminal
	cmd := exec.Command("go", "run", "../../../main.go", "policy", "create")
	cmd.Stdin = c.Tty()
	cmd.Stdout = c.Tty()
	cmd.Stderr = c.Tty()

	description := "This is my sample custom policy. There are many like it, but this one is mine."
	// Run through the wizard
	cmd.Start()
	expectAndRespond(assert, c, `Policies directory path:`, "test_custom_policies", 10)
	expectAndRespond(assert, c, `Create policies directory here`, "y", 1)
	defer os.RemoveAll("test_custom_policies")
	expectAndRespond(assert, c, `Select provider`, "AWS", 1)
	expectAndRespond(assert, c, `Select target`, "terraform", 1)
	expectAndRespond(assert, c, `Policy name`, "sample_custom_policy", 1)
	expectAndRespond(assert, c, `Title`, "My Sample Custom Policy", 1)
	expectAndRespond(assert, c, `Description`, description, 1)
	expectAndRespond(assert, c, `Category`, "Storage", 1)
	expectAndRespond(assert, c, `ResourceType`, "aws_s3_bucket", 1)
	expectAndRespond(assert, c, `Select severity`, "Info", 1)
	c.Tty().Close()
	out, err := c.ExpectEOF()

	// Test the result of the wizard
	createdPolicyDir := "test_custom_policies/policies/opal/sample_custom_policy"
	assert.Contains(out, "created:  "+createdPolicyDir)
	_, err = os.Stat(createdPolicyDir)
	assert.NoError(err)
	_, err = os.Stat(createdPolicyDir + "/metadata.yaml")
	assert.NoError(err)
	metadata, err := os.ReadFile(createdPolicyDir + "/metadata.yaml")
	expected := regexp.MustCompile("description: \"" + description + "\"")
	matches := expected.FindStringSubmatch(string(metadata))
	if matches == nil {
		assert.Fail("Created policy doesn't have expected description in metadata.yaml")
	}
	_, err = os.Stat(createdPolicyDir + "/terraform/policy.rego")
	assert.NoError(err)
	expected = regexp.MustCompile(`resource_type := "aws_s3_bucket"`)
	policy, err := os.ReadFile(createdPolicyDir + "/terraform/policy.rego")
	matches = expected.FindStringSubmatch(string(policy))
	if matches == nil {
		assert.Fail("Created policy doesn't have expected aws_s3_bucket in terraform/policy.rego")
	}
}

func expectAndRespond(assert *assert.Assertions, c *expect.Console, exp string, response string, timeout int) {
	_, err := c.Expect(expect.Regexp(regexp.MustCompile(exp)), expect.WithTimeout(time.Second*time.Duration(timeout)))
	assert.NoError(err)

	if response != "" {
		_, err = c.SendLine(response)
		assert.NoError(err)
	}
}
