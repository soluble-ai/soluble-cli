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
	"strings"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

func PartialDecodeJSON(buf []byte) map[string]string {
	r := map[string]string{}
	// gjson doesn't require valid JSON
	for k, v := range gjson.ParseBytes(buf).Map() {
		if v.Type == gjson.String {
			r[k] = v.Str
		}
	}
	return r
}

func PartialDecodeYAML(buf []byte) map[string]string {
	var m map[string]interface{}
	// truncated yaml is still mostly yaml, but we'll just ignore errors
	_ = yaml.Unmarshal(buf, &m)
	r := map[string]string{}
	for k, v := range m {
		if s, ok := v.(string); ok {
			r[k] = s
		}
	}
	return r
}

func PartialDecode(name string, buf []byte) map[string]string {
	switch {
	default:
		return PartialDecodeJSON(buf)
	case strings.HasSuffix(name, ".yaml"):
		return PartialDecodeYAML(buf)
	case strings.HasSuffix(name,".yml"):
		return PartialDecodeYAML(buf)
	}
}
