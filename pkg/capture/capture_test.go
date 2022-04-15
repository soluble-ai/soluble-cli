package capture

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCapture(t *testing.T) {
	testCapture(t, NewCapture(), "Hello world\n", []string{
		"Hello world",
	})
	cap := NewCapture()
	cap.MemoryLimit = 16
	testCapture(t, cap, "Hello world\none two three four five\n", []string{
		"Hello world",
		"one two three four five",
	})
}

func testCapture(t *testing.T, cap *Capture, expect string, lines []string) {
	assert := assert.New(t)
	assert.NotNil(cap)
	for _, line := range lines {
		fmt.Fprintln(cap, line)
	}
	r, err := cap.Output()
	assert.NoError(err)
	dat, err := io.ReadAll(r)
	assert.NoError(err)
	assert.Equal(expect, string(dat))
	assert.NoError(cap.Close())
}
