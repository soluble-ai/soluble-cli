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

func TestPathSupport(t *testing.T) {
	ps := &PathSupport{
		Path:   []string{"a", "b"},
		Limit:  2,
		SortBy: []string{"x"},
	}
	n := jnode.NewObjectNode()
	a := n.PutObject("a").PutArray("b")
	a.Append(jnode.NewObjectNode().Put("x", 1))
	a.Append(jnode.NewObjectNode().Put("x", 3))
	a.Append(jnode.NewObjectNode().Put("x", 2))
	rows := ps.GetRows(n)
	if len(rows) != 2 {
		t.Error(rows)
	}
	if rows[1].Path("x").AsInt() != 2 {
		t.Error(rows)
	}
}

func TestComputedColumns(t *testing.T) {
	count := 0
	ps := &PathSupport{
		Path: []string{"data"},
		ComputedColumns: map[string]ColumnFunction{
			"count": func(n *jnode.Node) interface{} {
				count++
				return count
			},
		},
	}
	n := jnode.NewObjectNode()
	data := n.PutArray("data")
	data.Append(jnode.NewObjectNode().Put("greeting", "hello"))
	data.Append(jnode.NewObjectNode().Put("greeting", "howdy"))
	rows := ps.GetRows(n)
	if len(rows) != 2 {
		t.Error(rows)
	}
	for i, row := range rows {
		if i+1 != row.Path("count").AsInt() || row.Path("greeting").IsMissing() {
			t.Error(row)
		}
	}
}
