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

func (p *VerticalPrinter) PrintResult(w io.Writer, result *jnode.Node) {
	rows := p.getRows(result)
	for i, row := range rows {
		if i > 0 {
			fmt.Fprintln(w)
		}
		for _, column := range p.Columns {
			fmt.Fprintf(w, "%s:", column)
			val := p.Format(column, row)
			if isMultiline(val) {
				fmt.Fprintf(w, "\n%s\n", val)
			} else {
				fmt.Fprintf(w, " %s\n", val)
			}
		}
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
