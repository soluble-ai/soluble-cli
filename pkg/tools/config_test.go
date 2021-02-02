package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	assert := assert.New(t)
	o := &ToolOpts{}
	assert.NotNil(o.GetConfig())
}
