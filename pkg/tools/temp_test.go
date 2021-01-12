package tools

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTempFile(t *testing.T) {
	assert := assert.New(t)
	f, err := TempFile("test*")
	assert.Nil(err)
	assert.NotEmpty(f)
	st, err := os.Stat(f)
	assert.Nil(err)
	assert.True(st.Mode().IsRegular())
	assert.Nil(os.Remove(f))
}
