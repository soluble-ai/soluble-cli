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
	assert.Len(t, m.Rules[CheckovYAML], 1)
	rule := m.Rules[CheckovYAML][0]
	assert.NotNil(t, rule)
	assert.True(t, strings.HasSuffix(rule.Path, "testdata/policies/checkov/team_tag"), rule.Path)
	assert.Equal(t, "c-ckv-team-tag", rule.ID)
	tmp, err := os.MkdirTemp("", "test*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmp)
	assert.NoError(t, m.PrepareRules(tmp))
	d, err := os.ReadFile(filepath.Join(tmp, "terraform-c-ckv-team-tag.yaml"))
	assert.NoError(t, err)
	fmt.Println(string(d))
	var ruleBody map[string]interface{}
	assert.NoError(t, yaml.Unmarshal(d, &ruleBody))
	ruleMetadata, _ := ruleBody["metadata"].(map[string]interface{})
	assert.NotNil(t, ruleMetadata)
	assert.Equal(t, "c-ckv-team-tag", ruleMetadata["id"])
	validate := m.ValidateRules()
	assert.NoError(t, validate.Errors)
}
