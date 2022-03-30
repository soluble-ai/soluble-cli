package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	assert := assert.New(t)
	m, err := Read("testdata/main.tf")
	assert.NoError(err)
	if assert.NotNil(m) {
		assert.Equal(1, m.ResourceCounts["aws_instance"])
		assert.Equal(2, m.ResourceCounts["aws_security_group"])
		if assert.NotNil(m.Settings) {
			assert.Equal("~> 1.0.0", m.Settings.RequiredVersion)
			assert.Equal("s3", m.Settings.Backend)
			assert.Contains(m.Settings.RequiredProviders, &RequiredProvider{
				Alias:   "aws",
				Source:  "hashicorp/aws",
				Version: "4.8.0",
			})
		}
		assert.Contains(m.ModulesUsed, &ModuleUse{
			Source:  "terraform-aws-modules/vpc/aws",
			Version: "3.14.0",
		})
	}
}
