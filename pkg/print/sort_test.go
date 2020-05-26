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
	"sort"
	"testing"

	"github.com/soluble-ai/go-jnode"
)

func TestSortBy(t *testing.T) {
	rows := []*jnode.Node{
		jnode.NewObjectNode().Put("one", "b").Put("two", "b"),
		jnode.NewObjectNode().Put("one", "b").Put("two", "a"),
		jnode.NewObjectNode().Put("one", "a").Put("two", "a"),
		jnode.NewObjectNode().Put("one", "a").Put("two", "b"),
		jnode.NewObjectNode().Put("one", "a").Put("two", "a").Put("three", "z"),
	}
	sort.Sort(&rowsSort{rows, []string{"one", "two", "three"}})
	assertEqual(t, rows[0], "a", "a", "")
	assertEqual(t, rows[1], "a", "a", "z")
	assertEqual(t, rows[2], "a", "b", "")
	assertEqual(t, rows[3], "b", "a", "")
	assertEqual(t, rows[4], "b", "b", "")
}

func assertEqual(t *testing.T, n *jnode.Node, one, two, three string) {
	if n.Path("one").AsText() != one || n.Path("two").AsText() != two ||
		n.Path("three").AsText() != three {
		t.Error(n)
	}
}
