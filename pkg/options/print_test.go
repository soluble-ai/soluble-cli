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
	"io"
	"testing"

	"github.com/soluble-ai/go-jnode"
)

func TestPrintResult(t *testing.T) {
	opts := &PrintOpts{}
	w := &bytes.Buffer{}
	opts.outputSource = func() io.Writer { return w }
	opts.PrintResult(jnode.NewObjectNode().Put("greeting", "hello"))
	if s := w.String(); s != "greeting: hello\n\n" {
		t.Error(s)
	}
}

func TestPrintLimit(t *testing.T) {
	opts := &PrintOpts{
		Limit:     1,
		Path:      []string{},
		Columns:   []string{"x"},
		NoHeaders: true,
	}
	w := &bytes.Buffer{}
	opts.outputSource = func() io.Writer { return w }
	opts.PrintResult(fromJSON(`[{"x":"one"},{"x":"two"}]`))
	if s := w.String(); s != "one\n" {
		t.Error(s)
	}
}

func fromJSON(s string) *jnode.Node {
	n, err := jnode.FromJSON([]byte(s))
	if err != nil {
		panic(err)
	}
	return n
}
