package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectPolicy(t *testing.T) {
	assert := assert.New(t)
	m, ruleType, rule, target, err := DetectPolicy("testdata")
	assert.NoError(err)
	assert.NotNil(m)
	assert.Nil(ruleType)
	assert.Nil(rule)
	assert.Empty(target)
	m, ruleType, rule, target, err = DetectPolicy("testdata/policies")
	assert.NoError(err)
	assert.NotNil(m)
	assert.Nil(ruleType)
	assert.Nil(rule)
	assert.Empty(target)
	m, ruleType, rule, target, err = DetectPolicy("testdata/policies/checkov")
	assert.NoError(err)
	assert.NotNil(m)
	assert.Equal(CheckovYAML, ruleType)
	assert.Nil(rule)
	assert.Empty(target)
	assert.Equal(1, len(m.Rules[CheckovYAML]))
	m, ruleType, rule, target, err = DetectPolicy("testdata/policies/checkov/team_tag")
	assert.NoError(err)
	assert.NotNil(m)
	assert.Equal(CheckovYAML, ruleType)
	if assert.NotNil(rule) {
		assert.Equal("c-ckv-team-tag", rule.ID)
	}
	assert.Empty(target)
	assert.Equal(1, len(m.Rules[CheckovYAML]))
	assert.Same(rule, m.Rules[CheckovYAML][0])
	m, ruleType, rule, target, err = DetectPolicy("testdata/policies/checkov/team_tag/terraform")
	assert.NoError(err)
	assert.NotNil(m)
	assert.Equal(CheckovYAML, ruleType)
	if assert.NotNil(rule) {
		assert.Equal("c-ckv-team-tag", rule.ID)
	}
	assert.Equal(Terraform, target)
	assert.Equal(1, len(m.Rules[CheckovYAML]))
	assert.Same(rule, m.Rules[CheckovYAML][0])
	m, ruleType, rule, target, err = DetectPolicy("testdata/policies/checkov/team_tag/terraform/tests")
	assert.NoError(err)
	assert.NotNil(m)
	assert.Equal(CheckovYAML, ruleType)
	if assert.NotNil(rule) {
		assert.Equal("c-ckv-team-tag", rule.ID)
	}
	assert.Equal(Terraform, target)
	assert.Equal(1, len(m.Rules[CheckovYAML]))
	assert.Same(rule, m.Rules[CheckovYAML][0])
}
