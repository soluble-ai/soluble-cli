package inventory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInventory(t *testing.T) {
	assert := assert.New(t)
	m := Do("testdata")
	if assert.NotNil(m) {
		assert.ElementsMatch(m.PythonDirectories.Values(), []string{"lang/python-app"})
		assert.ElementsMatch(m.JavaDirectories.Values(), []string{"lang/java", "lang/java2", "lang/java3"})
		assert.ElementsMatch(m.TerraformRootModules.Values(), []string{"tf/r1", "tf/r1j", "tf/r2"})
		assert.ElementsMatch(m.KustomizeDirectories.Values(), []string{"k/kus"})
		assert.ElementsMatch(m.KubernetesManifestDirectories.Values(), []string{"k/t"})
		assert.ElementsMatch(m.HelmCharts.Values(), []string{"k/h"})
	}
}
