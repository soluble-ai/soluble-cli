package util

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/soluble-ai/go-jnode"
)

func ReadJSONFile(filename string) (*jnode.Node, error) {
	path := filepath.FromSlash(filename)
	var r io.ReadCloser
	var err error
	r, err = os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			r, err = os.Open(path + ".gz")
			if err == nil {
				rr := r
				defer rr.Close()
				r, err = gzip.NewReader(r)
			}
		}
	}
	if err != nil {
		return nil, err
	}
	defer r.Close()
	d := json.NewDecoder(r)
	var n jnode.Node
	err = d.Decode(&n)
	return &n, err
}

func MustReadJSONFile(filename string) *jnode.Node {
	n, err := ReadJSONFile(filename)
	if err != nil {
		panic(err)
	}
	return n
}
