package checkov

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDockerfileDection(t *testing.T) {
	d := &Dockerfile{}
	d.Directory = "testdata/docker"
	assert := assert.New(t)
	assert.NoError(d.Validate())
	rel, err := filepath.Rel(d.GetDirectory(), d.dockerfile)
	assert.NoError(err)
	assert.Equal("Dockerfile", rel)

	d = &Dockerfile{}
	d.Dockerfile = "testdata/docker/build.dockerfile"
	assert.NoError(d.Validate())
	rel, err = filepath.Rel(d.GetDirectory(), d.dockerfile)
	assert.NoError(err)
	assert.Equal("build.dockerfile", rel)
}
