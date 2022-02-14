package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestCheckov(t *testing.T) {
	m := NewManager("testdata")
	err := m.LoadRules(CheckovYAML)
	assert.NoError(t, err)
	assert.Len(t, m.Rules[CheckovYAML], 1)
	rule := m.Rules[CheckovYAML][0]
	assert.NotNil(t, rule)
	assert.Equal(t, "testdata/policies/checkov/team_tag", rule.Path)
	assert.Equal(t, "c-ckv-team-tag", rule.ID)
	tmp, err := os.MkdirTemp("", "test*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmp)
	assert.NoError(t, m.PrepareRules(tmp, CheckovYAML, Terraform))
	d, err := os.ReadFile(filepath.Join(tmp, "terraform-c-ckv-team-tag.yaml"))
	assert.NoError(t, err)
	fmt.Println(string(d))
	var ruleBody map[string]interface{}
	assert.NoError(t, yaml.Unmarshal(d, &ruleBody))
	ruleMetadata, _ := ruleBody["metadata"].(map[string]interface{})
	assert.NotNil(t, ruleMetadata)
	assert.Equal(t, "c-ckv-team-tag", ruleMetadata["id"])
}
