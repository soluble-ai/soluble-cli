package print

import (
	"encoding/csv"
	"io"
	"sort"

	"github.com/soluble-ai/go-jnode"
)

type CSVPrinter struct {
	Filter
	NoHeaders  bool
	Path       []string
	Columns    []string
	SortBy     []string
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
