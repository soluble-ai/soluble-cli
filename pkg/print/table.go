package print

import (
	"bytes"
	"fmt"
	"io"
	"text/tabwriter"
	"unicode"

	"github.com/soluble-ai/go-jnode"
)

type TablePrinter struct {
	PathSupport
	Filter
	NoHeaders  bool
	Columns    []string
	Formatters Formatters
}

var _ Interface = &TablePrinter{}

func (p *TablePrinter) PrintResult(w io.Writer, result *jnode.Node) {
	tw := tabwriter.NewWriter(w, 5, 0, 1, ' ', 0)
	if !p.NoHeaders {
		p.PrintHeader(tw)
	}
	p.PrintRows(tw, result)
	_ = tw.Flush()
}

func (p *TablePrinter) PrintRows(w io.Writer, result *jnode.Node) {
	rows := p.getRows(result)
	for _, row := range rows {
		if !p.matches(row) {
			continue
		}
		for i, c := range p.Columns {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, p.Formatters.Format(c, row))
		}
		fmt.Fprint(w, "\n")
	}
}

func (p *TablePrinter) PrintHeader(w io.Writer) {
	for i, c := range p.Columns {
		if i > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, toHeader(c))
	}
	fmt.Fprint(w, "\n")
}

func toHeader(c string) string {
	w := &bytes.Buffer{}
	var wasUpper int
	if c[len(c)-1] == '+' {
		c = c[0 : len(c)-1]
	}
	for i, ch := range c {
		upper := unicode.IsUpper(ch)
		if i > 0 && wasUpper == 0 && upper {
			w.WriteRune('-')
		}
		if ch == '_' {
			w.WriteRune('-')
			wasUpper = -1
		} else {
			w.WriteRune(unicode.ToUpper(ch))
			wasUpper = 0
			if upper {
				wasUpper = 1
			}
		}
	}
	return w.String()
}
