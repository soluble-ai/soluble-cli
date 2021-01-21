package util

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/soluble-ai/go-jnode"
)

func ReadJSONFile(filename string) (*jnode.Node, error) {
	f, err := os.Open(filepath.FromSlash(filename))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	d := json.NewDecoder(f)
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
