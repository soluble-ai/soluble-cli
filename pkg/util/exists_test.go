package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	assert := assert.New(t)
	assert.True(FileExists("exists.go"))
	assert.False(FileExists("foo/nope.go"))
	assert.True(DirExists("../util"))
	assert.False(DirExists("foo"))
	assert.False(DirEmpty("."))
	d, err := os.MkdirTemp("", "test*")
	if assert.NoError(err) {
		defer os.RemoveAll(d)
		assert.True(DirExists(d))
		assert.True(DirEmpty(d))
	}
}
