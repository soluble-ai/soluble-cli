package tools

import (
	"testing"

	"github.com/soluble-ai/go-jnode"
	"github.com/stretchr/testify/assert"
)

func TestPassFormatter(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("PASS", PassFormatter(jnode.NewObjectNode().Put("pass", true).Path("pass")))
}
