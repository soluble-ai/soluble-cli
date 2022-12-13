package inventory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	assert := assert.New(t)
	d, err := readContent("testdata/example.yaml", make([]byte, 8192))
	if assert.NoError(err) {
		m := d.DecodeDocument()
		assert.Contains(m, "greeting")
	}
}
