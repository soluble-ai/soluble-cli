package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSize(t *testing.T) {
	for _, tc := range []struct {
		n uint64
		e string
	}{
		{0, "0B"},
		{1000, "1000B"},
		{1592, "1K"},
		{1024*1024*5 + 100, "5M"},
		{1024*1024*1024*2 + 1000, "2G"},
	} {
		assert.Equal(t, tc.e, Size(tc.n))
	}
}
