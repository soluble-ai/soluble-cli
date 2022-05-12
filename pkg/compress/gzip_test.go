package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGZIPPipe(t *testing.T) {
	assert := assert.New(t)
	dat, err := os.ReadFile("gzip.go")
	if !assert.NoError(err) {
		return
	}
	gz := NewGZIPPipe(bytes.NewReader(dat))
	gzdat := &bytes.Buffer{}
	if _, err := io.Copy(gzdat, gz); !assert.NoError(err) {
		return
	}
	assert.NoError(gz.Close())
	gzr, err := gzip.NewReader(bytes.NewReader(gzdat.Bytes()))
	if !assert.NoError(err) {
		return
	}
	rt, err := io.ReadAll(gzr)
	if assert.NoError(err) {
		assert.Equal(dat, rt)
	}
}
