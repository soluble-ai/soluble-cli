package checkov

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/policy/manager"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestCheckov(t *testing.T) {
	m := &manager.M{}
	err := m.DetectPolicy("testdata")
	assert.NoError(t, err)
	assert.Len(t, m.Policies[CheckovYAML], 1)
	policy := m.Policies[CheckovYAML][0]
	assert.NotNil(t, policy)
	assert.True(t, strings.HasSuffix(policy.Path, "testdata/policies/checkov/team_tag"), policy.Path)
	assert.Equal(t, "c-ckv-team-tag", policy.ID)
	tmp, err := os.MkdirTemp("", "test*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmp)
	assert.NoError(t, m.PreparePolicies(tmp))
	d, err := os.ReadFile(filepath.Join(tmp, "terraform-c-ckv-team-tag.yaml"))
	assert.NoError(t, err)
	fmt.Println(string(d))
	var policyBody map[string]interface{}
	assert.NoError(t, yaml.Unmarshal(d, &policyBody))
	policyMetadata, _ := policyBody["metadata"].(map[string]interface{})
	assert.NotNil(t, policyMetadata)
	assert.Equal(t, "c-ckv-team-tag", policyMetadata["id"])
	validate := m.ValidatePolicies()
	assert.NoError(t, validate.Errors)
}
