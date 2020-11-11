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

	"github.com/soluble-ai/go-jnode"
)

type CSVPrinter struct {
	PathSupport
	NoHeaders  bool
	Columns    []string
	Formatters Formatters
}

var _ Interface = &CSVPrinter{}

func (p *CSVPrinter) PrintResult(w io.Writer, result *jnode.Node) int {
	cw := csv.NewWriter(w)
	if !p.NoHeaders {
		p.printHeaders(cw)
	}
	defer cw.Flush()
	return p.printRows(cw, result)
}

func (p *CSVPrinter) printHeaders(w *csv.Writer) {
	row := make([]string, len(p.Columns))
	for i, c := range p.Columns {
		row[i] = toHeader(c)
	}
	_ = w.Write(row)
}

func (p *CSVPrinter) printRows(w *csv.Writer, result *jnode.Node) int {
	rows := p.getRows(result)
	for _, row := range rows {
		rec := make([]string, len(p.Columns))
		for i, c := range p.Columns {
			rec[i] = p.Formatters.Format(c, row)
		}
		_ = w.Write(rec)
	}
	return len(rows)
}
