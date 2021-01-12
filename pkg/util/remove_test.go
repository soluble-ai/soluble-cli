package util

import (
	"testing"

	"github.com/soluble-ai/go-jnode"
	"github.com/stretchr/testify/assert"
)

func TestRemoveJNodeElementsIf(t *testing.T) {
	assert := assert.New(t)
	n := jnode.NewArrayNode().Append("hello").
		Append("world").
		Append("one").
		Append("two").
		Append("three")
	m := RemoveJNodeElementsIf(n, func(*jnode.Node) bool { return false })
	assert.Same(n, m)
	l := RemoveJNodeElementsIf(n, func(e *jnode.Node) bool { return e.AsText() == "two" })
	assert.Equal(4, l.Size())
	assert.Equal("hello", l.Get(0).AsText())
	assert.Equal("three", l.Get(3).AsText())
}
