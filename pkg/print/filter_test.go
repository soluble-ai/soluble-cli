// Copyright 2020 Soluble Inc
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

package print

import (
	"testing"

	"github.com/soluble-ai/go-jnode"
)

func TestFilter(t *testing.T) {
	row := jnode.NewObjectNode().Put("name", "value").Put("greeting", "hello").Put("fail", true)
	if n := NewSingleFilter("hello").(*singleFilter); n.name != "" || n.g == nil || !n.Matches(row) {
		t.Error(n)
	}
	if n := NewSingleFilter("name=").(*singleFilter); n.name != "name" || !n.Matches(row) {
		t.Error(n)
	}
	if n := NewSingleFilter(""); !n.Matches(row) {
		t.Error(n)
	}
	if n := NewSingleFilter("world"); n.Matches(row) {
		t.Error(n)
	}
	if n := NewSingleFilter("name=joe*"); n.Matches(row) {
		t.Error(n)
	}
	if n := NewSingleFilter("name=v*"); !n.Matches(row) {
		t.Error(n)
	}
	if n := NewSingleFilter("name!=value"); n.Matches(row) {
		t.Error(n)
	}
	if n := NewAndFilter(nil); !n.Matches(row) {
		t.Error(n)
	}
	if n := NewAndFilter([]string{"name=value", "greeting=hello"}); !n.Matches(row) {
		t.Error(n)
	}
	if n := NewSingleFilter("pass=false"); !n.Matches(row) {
		t.Error(n)
	}
	if n := NewSingleFilter("fail=true"); !n.Matches(row) {
		t.Error(n)
	}
}
