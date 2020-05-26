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
	"testing"

	"github.com/soluble-ai/go-jnode"
)

func TestValue(t *testing.T) {
	vp := NewValuePrinter("value(x.y)", []string{"data"}, nil)
	n := jnode.NewObjectNode()
	n.PutObject("a").Put("b", "hello")
	a := n.PutArray("data")
	a.AppendObject().PutObject("x").Put("y", 1)
	w := &bytes.Buffer{}
	vp.PrintResult(w, n)
	if s := w.String(); s != "1\n" {
		t.Error(s)
	}
	vp = NewValuePrinter("value(a.b)", nil, nil)
	w.Reset()
	vp.PrintResult(w, n)
	if s := w.String(); s != "hello\n" {
		t.Error(s)
	}
}
