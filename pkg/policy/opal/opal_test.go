package opal

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/policy/manager"
	"github.com/stretchr/testify/assert"
)

func TestPolicies(t *testing.T) {
	assert := assert.New(t)
	m := &manager.M{}
	err := m.DetectPolicy("testdata/policies")
	assert.NoError(err)
	assert.Equal(1, len(m.Policies))
	tm, err := m.TestPolicies()
	assert.NoError(err)
	assert.Equal(0, tm.Failed)
}
