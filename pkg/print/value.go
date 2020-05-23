package print

import (
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type ValuePrinter struct {
	PathSupport
	Name string
}

var _ Interface = &ValuePrinter{}

var valueSpecRe = regexp.MustCompile(`value\((.*)\)`)

func NewValuePrinter(format string, path []string, sortBy []string) *ValuePrinter {
	m := valueSpecRe.FindStringSubmatch(format)
	if m[1] == "" {
		log.Warnf("invalid value specifier {warning:%s} - must be in the form 'value(name)'", format)
		os.Exit(2)
	}
	return &ValuePrinter{
		PathSupport: PathSupport{Path: path, SortBy: sortBy},
		Name:        m[1],
	}
}

func (p *ValuePrinter) PrintResult(w io.Writer, result *jnode.Node) {
	if len(p.Path) == 0 {
		n := result.Path(p.Name)
		if !n.IsMissing() {
			fmt.Fprintln(w, n.AsText())
		}
	} else {
		for _, row := range p.getRows(result) {
			n := row.Path(p.Name)
			if !n.IsMissing() {
				fmt.Fprintln(w, n.AsText())
			}
		}
	}
}
