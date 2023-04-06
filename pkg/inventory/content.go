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

package inventory

import (
	"fmt"
	"os"
	"strings"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

type Content struct {
	path string
	Head []byte
	doc  map[string]string
}

func readContent(path string, buf []byte) (*Content, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	n, err := f.Read(buf)
	if n != 0 && err != nil {
		return nil, err
	}
	return &Content{
		path: path,
		Head: buf[0:n],
	}, nil
}

func decodeJSON(buf []byte) map[string]string {
	r := map[string]string{}
	// gjson doesn't require valid JSON
	for k, v := range gjson.ParseBytes(buf).Map() {
		if v.Type == gjson.String {
			r[k] = v.Str
		}
	}
	return r
}

func decodeYAML(buf []byte) map[string]string {
	var m map[string]interface{}
	// truncated yaml is still mostly yaml, but we'll just ignore errors
	_ = yaml.Unmarshal(buf, &m)
	r := map[string]string{}
	for k, v := range m {
		r[k] = fmt.Sprintf("%v", v)
	}
	return r
}

func (c *Content) DecodeDocument() map[string]string {
	if c.doc == nil {
		switch {
		case strings.HasSuffix(c.path, ".yaml") || strings.HasSuffix(c.path, ".yml"):
			c.doc = decodeYAML(c.Head)
		default:
			c.doc = decodeJSON(c.Head)
		}
	}
	return c.doc
}
