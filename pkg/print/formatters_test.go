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
	"time"

	"github.com/soluble-ai/go-jnode"
)

func TestFormatters(t *testing.T) {
	var f Formatters
	if s := f.Format("s", jnode.NewObjectNode().Put("s", "hello")); s != "hello" {
		t.Error(s)
	}
}

func TestTimestampFormatters(t *testing.T) {
	now := time.Date(2020, 5, 11, 10, 5, 45, 0, time.UTC)
	formatterNow = &now
	formatterLocation = time.FixedZone("test", -8*60*60)
	n := jnode.NewObjectNode().Put("ts", "2020-05-08T11:18:33Z")
	if s := TimestampFormatter(n.Path("ts")); s != "2020-05-08T03:18:33-08:00" {
		t.Error("timestamp wrong", n, s)
	}
	if s := RelativeTimestampFormatter(n.Path("ts")); s != "2d22h47m12s" {
		t.Error("relative ts wrong", n, s)
	}
	n = jnode.NewObjectNode().Put("ts", "9999-12-31T15:59:59-08:00")
	if s := RelativeTimestampFormatter(n.Path("ts")); s != ">100y" {
		t.Error("long relative time is wrong", n, s)
	}
}

var bytesTestCases = []struct {
	value    int
	expected string
}{
	{10, "10"},
	{1536, "1.5K"},
	{1258291, "1.2M"},
	{2684354560, "2.5G"},
}

func TestBytesFormatter(t *testing.T) {
	for _, c := range bytesTestCases {
		n := jnode.NewObjectNode().Put("n", c.value)
		s := BytesFormatter(n.Path("n"))
		if s != c.expected {
			t.Error(c.value, c.expected, s)
		}
	}
}

var durationTestCases = []struct {
	value    int
	expected string
}{
	{1000, "1s"},
	{30 * 1000, "30s"},
	{65 * 1000, "1m5s"},
	{150, "150ms"},
}

func TestDurationFormatter(t *testing.T) {
	for _, c := range durationTestCases {
		n := jnode.NewObjectNode().Put("millis", c.value)
		s := DurationMillisFormatter(n.Path("millis"))
		if s != c.expected {
			t.Error(c.value, c.expected, s)
		}
	}
}

func TestGetCellValue(t *testing.T) {
	var testCases = []struct {
		n *jnode.Node
		c string
		v string
	}{
		{makeNode("1", "one"), "one", "1"},
		{makeNode("2", "one", "two"), "one.two", "2"},
		{makeNode("3", "one", "two", "three.dot"), "one.two.three..dot", "3"},
		{makeNode("4", "one.dot"), "one..dot", "4"},
	}
	for _, c := range testCases {
		v := getCellValue(c.n, c.c).AsText()
		if v != c.v {
			t.Error(c.n, c.c, c.v, v)
		}
	}
}

func makeNode(val string, path ...string) *jnode.Node {
	n := jnode.NewObjectNode()
	x := n
	for _, p := range path[0 : len(path)-1] {
		x = x.PutObject(p)
	}
	x.Put(path[len(path)-1], val)
	return n
}
