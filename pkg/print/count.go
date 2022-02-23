package print

import (
	"fmt"
	"io"

	"github.com/soluble-ai/go-jnode"
)

type CountPrinter struct {
	PathSupport
}

var _ Interface = &CountPrinter{}

func (p *CountPrinter) PrintResult(w io.Writer, result *jnode.Node) int {
	rows := p.GetRows(result)
	n := len(rows)
	fmt.Fprintf(w, "%d\n", n)
	return n
}
