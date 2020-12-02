package inventory

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocker(t *testing.T) {
	assert := assert.New(t)
	m := &Manifest{}
	m.scan("testdata", dockerDetector(0))
	assert.ElementsMatch(m.DockerDirectories.Values(),
		[]string{filepath.FromSlash("d/dot"), filepath.FromSlash("d/simple"), filepath.FromSlash("d/rdot")})
}
