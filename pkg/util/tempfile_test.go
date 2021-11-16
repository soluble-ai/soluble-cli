package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTempFile(t *testing.T) {
	assert := assert.New(t)
	path, err := TempFile("testing")
	if !assert.NoError(err) {
		return
	}
	assert.NoError(os.Remove(path))
}
