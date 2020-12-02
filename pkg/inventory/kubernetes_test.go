package inventory

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelm(t *testing.T) {
	assert := assert.New(t)
	m := &Manifest{}
	m.scan(filepath.Join("testdata", "k"), kubernetesDetector(0))
	assert.ElementsMatch(m.HelmCharts.Values(), []string{
		"h",
	})
	assert.ElementsMatch(m.KubernetesManifestDirectories.Values(), []string{
		"t",
	})
}
