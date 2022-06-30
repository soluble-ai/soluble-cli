package checkov

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/stretchr/testify/assert"
)

func TestPreparePy(t *testing.T) {
	assert := assert.New(t)
	m := &policy.Manager{Dir: "testdata"}
	err := m.DetectPolicy()
	if !assert.NoError(err) {
		return
	}
	if assert.Len(m.Rules[CheckovPython], 1) {
		rule := m.Rules[CheckovPython][0]
		assert.NotNil(rule)
	}
	temp, err := os.MkdirTemp("", "checkov-py-test")
	if !assert.NoError(err) {
		return
	}
	defer os.RemoveAll(temp)
	assert.NoError(m.PrepareRules(temp))
	dat, err := os.ReadFile(filepath.Join(temp, "c-ckv-py-s3-naming-terraform.py"))
	assert.NoError(err)
	fmt.Println(string(dat))
}
