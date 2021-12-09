package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	assert := assert.New(t)
	assert.True(FileExists("exists.go"))
	assert.False(FileExists("foo/nope.go"))
}
