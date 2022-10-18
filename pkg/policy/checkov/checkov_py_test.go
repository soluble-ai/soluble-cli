package checkov

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/policy/manager"
	"github.com/stretchr/testify/assert"
)

func TestPreparePy(t *testing.T) {
	assert := assert.New(t)
	m := &manager.M{}
	err := m.DetectPolicy("testdata/policies/checkov-py")
	if !assert.NoError(err) {
		return
	}
	if assert.Len(m.Policies[CheckovPython], 1) {
		policy := m.Policies[CheckovPython][0]
		assert.NotNil(policy)
	}
	temp, err := os.MkdirTemp("", "checkov-py-test")
	if !assert.NoError(err) {
		return
	}
	defer os.RemoveAll(temp)
	assert.NoError(m.PreparePolicies(temp))
	dat, err := os.ReadFile(filepath.Join(temp, "c-ckvpy-s3-naming-terraform.py"))
	assert.NoError(err)
	fmt.Println(string(dat))
	validate := m.ValidatePolicies()
	assert.NoError(validate.Errors)
	assert.Equal(1, validate.Valid)
	assert.Equal(0, validate.Invalid)
}
