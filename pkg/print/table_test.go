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
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/soluble-ai/go-jnode"
)

func TestHeader(t *testing.T) {
	if s := toHeader("updateTs"); s != "UPDATE-TS" {
		t.Error(s)
	}
	if s := toHeader("update_Ts"); s != "UPDATE-TS" {
		t.Error(s)
	}
	if s := toHeader("rule.id"); s != "RULE.ID" {
		t.Error(s)
	}
	if s := toHeader("Rule.Id"); s != "RULE.ID" {
		t.Error(s)
	}
}

func TestTable(t *testing.T) {
	w := &bytes.Buffer{}
	printer := &TablePrinter{
		PathSupport: PathSupport{Path: []string{"results"}},
		Columns:     []string{"name", "value"},
	}
	n := jnode.NewObjectNode()
	n.PutArray("results").Append(nv("greeting", "hello")).
		Append(nv("subject", "world"))
	printer.PrintResult(w, n)
	if s := w.String(); !strings.Contains(s, "NAME") || !strings.Contains(s, "world") {
		t.Error(s)
	}
}

func nv(n, v string) *jnode.Node {
	return jnode.NewObjectNode().Put("name", n).Put("value", v)
}

func TestFormat(t *testing.T) {
	printer := &TablePrinter{
		PathSupport: PathSupport{Path: []string{"rows"}},
		Columns:     []string{"v1", "v2"},
		Formatters: map[string]Formatter{
			"v1": func(n *jnode.Node) string {
				return "xxx"
			},
		},
	}
	n := jnode.NewObjectNode()
	a := n.PutArray("rows")
	a.Append(jnode.NewObjectNode().Put("v1", "A").Put("v2", "B"))
	a.Append(jnode.NewObjectNode().Put("v1", "C").Put("v2", "D"))
	w := &bytes.Buffer{}
	printer.PrintResult(w, n)
	s := w.String()
	if strings.Contains(s, "A") || strings.Contains(s, "C") || !strings.Contains(s, "xxx") {
		t.Error("formatting failed", s)
	}
}

func TestTs(t *testing.T) {
	now := time.Now()
	formatterNow = &now
	opts := TablePrinter{
		PathSupport: PathSupport{Path: []string{"rows"}},
		Columns:     []string{"updateTs+", "createTs"},
	}
	n := jnode.NewObjectNode()
	a := n.PutArray("rows")
	a.Append(jnode.NewObjectNode().
		Put("updateTs", now.Add(-4500*time.Millisecond).Format(time.RFC3339)).
		Put("createTs", "2020-03-23T16:36:14-07:00"))
	w := &bytes.Buffer{}
	opts.PrintResult(w, n)
	s := w.String()
	if !strings.Contains(s, "5s") || !strings.Contains(s, "2020-03-23T") {
		t.Error(n, s)
	}
}
