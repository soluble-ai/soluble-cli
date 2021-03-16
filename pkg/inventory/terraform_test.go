package inventory

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerraform(t *testing.T) {
	assert := assert.New(t)
	m := &Manifest{}
	m.scan(filepath.Join("testdata", "tf"), &terraformDetector{})
	assert.ElementsMatch(m.TerraformRootModules.Values(), []string{
		"r1", "r1j", "r2",
	})
	assert.ElementsMatch(m.TerraformModules.Values(), []string{
		"r1", "r1j", "r2", "m1",
	})
}
