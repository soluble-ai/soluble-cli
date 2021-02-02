package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	assert := assert.New(t)
	fill := func(s string) interface{} { return s }
	c := NewCache(2)
	assert.Equal("hello", c.Get("hello", fill))
	assert.Equal("world", c.Get("world", fill))
	assert.Equal("hello", c.Get("hello", fill))
	assert.Equal(uint32(2), c.entries["hello"].use)
	assert.Equal(uint32(1), c.entries["world"].use)
	c.Put("blah", 5)
	assert.Equal(5, c.Get("blah", fill))
	assert.False(c.Contains("world"))
	assert.True(c.Contains("hello"))
}
