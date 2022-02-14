package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenericSet(t *testing.T) {
	var m map[string]interface{}
	GenericSet(&m, "metadata/id", "hello")
	assert.NotNil(t, m)
	md, ok := m["metadata"]
	assert.True(t, ok)
	mdm, ok := md.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "hello", mdm["id"])
	assert.Equal(t, "hello", GenericGet(m, "metadata/id"))
}
