package terraform

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLatest(t *testing.T) {
	assert := assert.New(t)
	data, err := getTestData()
	if err != nil {
		assert.Fail(err.Error())
	}
	v := parseLatestVersion(bytes.NewBuffer(data))
	assert.Equal("0.14.10", v)
}

func getTestData() ([]byte, error) {
	r, err := os.Open("testdata/releases.html.gz")
	if err != nil {
		return nil, err
	}
	defer r.Close()
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()
	return io.ReadAll(gzr)
}
