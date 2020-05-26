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
	"encoding/csv"
	"io"
	"sort"

	"github.com/soluble-ai/go-jnode"
)

type CSVPrinter struct {
	PathSupport
	Filter
	NoHeaders  bool
	Columns    []string
	Formatters Formatters
}

var _ Interface = &CSVPrinter{}

func (p *CSVPrinter) PrintResult(w io.Writer, result *jnode.Node) {
	cw := csv.NewWriter(w)
	if !p.NoHeaders {
		p.printHeaders(cw)
	}
	p.printRows(cw, result)
	cw.Flush()
}

func (p *CSVPrinter) printHeaders(w *csv.Writer) {
	row := make([]string, len(p.Columns))
	for i, c := range p.Columns {
		row[i] = toHeader(c)
	}
	_ = w.Write(row)
}

func (p *CSVPrinter) printRows(w *csv.Writer, result *jnode.Node) {
	r := result
	for _, p := range p.Path {
		r = r.Path(p)
	}
	rows := r.Elements()
	if p.SortBy != nil {
		sort.Sort(&rowsSort{rows, p.SortBy})
	}
	for _, row := range rows {
		if !p.matches(row) {
			continue
		}
		rec := make([]string, len(p.Columns))
		for i, c := range p.Columns {
			rec[i] = p.Formatters.Format(c, row)
		}
		_ = w.Write(rec)
	}
}
