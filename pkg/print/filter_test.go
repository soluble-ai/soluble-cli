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
	row := jnode.NewObjectNode().Put("name", "value").Put("greeting", "hello")
	if n := NewFilter("hello"); n.name != "" || n.g == nil || !n.matches(row) {
		t.Error(n)
	}
	if n := NewFilter("name="); n.name != "name" || !n.matches(row) {
		t.Error(n)
	}
	if n := NewFilter(""); !n.matches(row) {
		t.Error(n)
	}
	if n := NewFilter("world"); n.matches(row) {
		t.Error(n)
	}
	if n := NewFilter("name=joe*"); n.matches(row) {
		t.Error(n)
	}
	if n := NewFilter("name=v*"); !n.matches(row) {
		t.Error(n)
	}
	if n := NewFilter("name!=value"); n.matches(row) {
		t.Error(n)
	}
}
