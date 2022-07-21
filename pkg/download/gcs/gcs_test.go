package gcs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {
	r := NewResolver("soluble-public", "tfscore")
	version, url, err := r("latest")
	assert := assert.New(t)
	assert.NoError(err)
	assert.True(strings.HasPrefix(url, "https://"), url)
	assert.NotEmpty(version)
}
