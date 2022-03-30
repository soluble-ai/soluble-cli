package repotree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInventory(t *testing.T) {
	assert := assert.New(t)
	root, err := FindRepoRoot(".")
	assert.NoError(err)
	tree, err := Do(root)
	assert.NoError(err)
	if !assert.NotNil(tree) {
		return
	}
	assert.Greater(tree.FileCount, 300)
	assert.Contains(tree.TerraformBackends, "s3")
	assert.NotNil(tree.GetFile("go.sum"))
	f := tree.GetFile("pkg/repotree/terraform/testdata/main.tf")
	if assert.NotNil(f) {
		assert.Equal("pkg/repotree/terraform/testdata/main.tf", f.Path)
		if assert.NotNil(f.Terraform) {
			if assert.NotNil(f.Terraform.Settings) {
				assert.Equal("~> 1.0.0", f.Terraform.Settings.RequiredVersion)
			}
		}
	}
	use := tree.TerraformExternalModules["terraform-aws-modules/security-group/aws"]
	assert.Equal("4.9.0", use.Version)
	assert.Equal(1, use.UsageCount)
}
