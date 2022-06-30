package checkov

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/stretchr/testify/assert"
)

func TestDetectPolicy(t *testing.T) {
	for _, dir := range []string{
		"testdata", "testdata/policies", "testdata/policies/checkov",
		"testdata/policies/checkov/team_tag/terraform",
	} {
		m := &policy.Manager{Dir: dir}
		err := m.DetectPolicy()
		assertFoundRule(t, m, dir, err)
	}
}

func assertFoundRule(t *testing.T, m *policy.Manager, dir string, err error) {
	t.Helper()
	assert := assert.New(t)
	assert.NoError(err)
	if assert.NotNil(m) && assert.Len(m.Rules[CheckovYAML], 1, dir) {
		rule := m.Rules[CheckovYAML][0]
		assert.ElementsMatch(rule.Targets, []policy.Target{policy.Terraform})
	}
}
