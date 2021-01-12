package root

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrap(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(`here is
  some text
  longer
  than 10
  chars`, wrap(2, 11, "here is some text longer than 10 chars"))
}
