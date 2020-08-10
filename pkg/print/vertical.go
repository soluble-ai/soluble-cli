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
	width := getLabelWidth(p.Columns)
	for i, row := range rows {
		if i > 0 {
			fmt.Fprintln(w)
		}
		for _, column := range p.Columns {
			printLabel(w, column, width, p.Format(column, row))
		}
	}
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
