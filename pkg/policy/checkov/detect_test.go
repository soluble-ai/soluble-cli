package checkov

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/stretchr/testify/assert"
)

func TestDetectPolicy(t *testing.T) {
	m, err := policy.DetectPolicy("testdata")
	assertFoundRule(t, m, err)
	m, err = policy.DetectPolicy("testdata/policies")
	assertFoundRule(t, m, err)
	m, err = policy.DetectPolicy("testdata/policies/checkov")
	assertFoundRule(t, m, err)
	m, err = policy.DetectPolicy("testdata/policies/checkov/team_tag/terraform")
	assertFoundRule(t, m, err)
}

func assertFoundRule(t *testing.T, m *policy.Manager, err error) {
	t.Helper()
	assert := assert.New(t)
	assert.NoError(err)
	if assert.NotNil(m) && assert.Len(m.Rules[CheckovYAML], 1) {
		rule := m.Rules[CheckovYAML][0]
		assert.ElementsMatch(rule.Targets, []policy.Target{policy.Terraform})
	}
}
