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

package print

import (
	"fmt"
	"io"

	"github.com/soluble-ai/go-jnode"
)

type VerticalPrinter struct {
	PathSupport
	Columns []string
	Formatters
}

var _ Interface = &VerticalPrinter{}

func (p *VerticalPrinter) PrintResult(w io.Writer, result *jnode.Node) int {
	rows := p.GetRows(result)
	width := getLabelWidth(p.Columns)
	for i, row := range rows {
		if i > 0 {
			fmt.Fprintln(w)
		}
		for _, column := range p.Columns {
			printLabel(w, column, width, p.Format(column, row))
		}
	}
	return len(rows)
}

func getLabelWidth(columnNames []string) int {
	m := 1
	for _, name := range columnNames {
		if l := len(name) + 1; l > m {
			m = l
		}
	}
	return m
}

func printLabel(w io.Writer, columnName string, labelWidth int, val string) {
	if isMultiline(val) {
		fmt.Fprintf(w, "%s:\n%s\n", columnName, val)
	} else {
		fmt.Fprintf(w, fmt.Sprintf("%%-%ds %%s\n", labelWidth), fmt.Sprintf("%s:", columnName),
			val)
	}
}

func isMultiline(text string) bool {
	n := 0
	for _, ch := range text {
		if ch == '\n' {
			n++
			if n > 2 {
				return true
			}
		}
	}
	return false
}
