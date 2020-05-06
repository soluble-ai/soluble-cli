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

package options

import (
	"bytes"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/soluble-ai/go-jnode"
)

func TestPrintResult(t *testing.T) {
	opts := &PrintOpts{}
	w := &bytes.Buffer{}
	opts.Output = w
	opts.PrintResult(jnode.NewObjectNode().Put("greeting", "hello"))
	if s := w.String(); s != "greeting: hello\n\n" {
		t.Error(s)
	}
	w.Reset()
	opts.Path = []string{"results"}
	opts.Columns = []string{"name", "value"}
	n := jnode.NewObjectNode()
	n.PutArray("results").Append(nv("greeting", "hello")).
		Append(nv("subject", "world"))
	opts.PrintResult(n)
	if s := w.String(); !strings.Contains(s, "NAME") || !strings.Contains(s, "world") {
		t.Error(s)
	}
}

func nv(n, v string) *jnode.Node {
	return jnode.NewObjectNode().Put("name", n).Put("value", v)
}

func TestHeader(t *testing.T) {
	if s := toHeader("updateTs"); s != "UPDATE-TS" {
		t.Error(s)
	}
	if s := toHeader("update_Ts"); s != "UPDATE-TS" {
		t.Error(s)
	}
}

func TestFormat(t *testing.T) {
	opts := PrintOpts{
		Path:    []string{"rows"},
		Columns: []string{"v1", "v2"},
	}
	opts.SetFormatter("v1", func(n *jnode.Node, columnName string) string {
		return "xxx"
	})
	n := jnode.NewObjectNode()
	a := n.PutArray("rows")
	a.Append(jnode.NewObjectNode().Put("v1", "A").Put("v2", "B"))
	a.Append(jnode.NewObjectNode().Put("v1", "C").Put("v2", "D"))
	w := &bytes.Buffer{}
	opts.Output = w
	opts.PrintResult(n)
	s := w.String()
	if strings.Contains(s, "A") || strings.Contains(s, "C") || !strings.Contains(s, "xxx") {
		t.Error("formatting failed", s)
	}
}

func TestTs(t *testing.T) {
	now := time.Now()
	opts := PrintOpts{
		Path:    []string{"rows"},
		Columns: []string{"updateTs+", "createTs"},
		now:     &now,
	}
	n := jnode.NewObjectNode()
	a := n.PutArray("rows")
	a.Append(jnode.NewObjectNode().
		Put("updateTs", now.Add(-4500*time.Millisecond).Format(time.RFC3339)).
		Put("createTs", "2020-03-23T16:36:14-07:00"))
	w := &bytes.Buffer{}
	opts.Output = w
	opts.PrintResult(n)
	s := w.String()
	if !strings.Contains(s, "5s") || !strings.Contains(s, "2020-03-23T") {
		t.Error(s)
	}
}

func TestTransformer(t *testing.T) {
	opts := PrintOpts{
		Path:    []string{"rows"},
		Columns: []string{"y"},
	}
	opts.AddTransformer(func(n *jnode.Node) *jnode.Node {
		n.Put("y", n.Path("x").AsText())
		return n
	})
	n := jnode.NewObjectNode()
	a := n.PutArray("rows")
	a.Append(jnode.NewObjectNode().
		Put("x", "1"))
	w := &bytes.Buffer{}
	opts.Output = w
	opts.PrintResult(n)
	s := w.String()
	if !strings.Contains(s, "1") {
		t.Error(s)
	}
}

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
