package checkov

import (
	"testing"

	policies "github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/policy/manager"
	"github.com/stretchr/testify/assert"
)

func TestDetectPolicy(t *testing.T) {
	for _, dir := range []string{
		"testdata", "testdata/policies", "testdata/policies/checkov",
		"testdata/policies/checkov/team_tag/terraform",
	} {
		m := &manager.M{}
		err := m.DetectPolicy(dir)
		assertFoundPolicy(t, m, dir, err)
	}
}

func assertFoundPolicy(t *testing.T, m *manager.M, dir string, err error) {
	t.Helper()
	assert := assert.New(t)
	assert.NoError(err)
	if assert.NotNil(m) && assert.Len(m.Policies[CheckovYAML], 1, dir) {
		policy := m.Policies[CheckovYAML][0]
		assert.ElementsMatch(policy.Targets, []policies.Target{policies.Terraform})
	}
}
