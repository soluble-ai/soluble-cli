// Copyright 2021 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/go-jnode"
)

func ReadJSONFile(filename string) (*jnode.Node, error) {
	path := filepath.FromSlash(filename)
	var r io.ReadCloser
	var err error
	r, err = os.Open(path)
	if err == nil && strings.HasSuffix(filename, ".gz") {
		rr := r
		defer rr.Close()
		r, err = gzip.NewReader(r)
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
