package checkov

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuote(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("hello there", quote("hello there"))
	assert.Equal(`hello \"there\"`, quote(`hello "there"`))
}
