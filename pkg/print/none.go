package print

import (
	"io"

	"github.com/soluble-ai/go-jnode"
)

type NonePrinter struct{}

var _ Interface = &NonePrinter{}

func (p *NonePrinter) PrintResult(w io.Writer, result *jnode.Node) {}
