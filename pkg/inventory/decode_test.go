package inventory

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	assert := assert.New(t)
	d, err := os.ReadFile("testdata/example.yaml")
	if assert.NoError(err) {
		m := decodeDocument("example.yaml", d)
		assert.Contains(m, "greeting")
	}
}
