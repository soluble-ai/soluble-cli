package print

import (
	"encoding/json"
	"io"

	"github.com/soluble-ai/go-jnode"
)

type JSONPrinter struct{}

var _ Interface = &JSONPrinter{}

func (p *JSONPrinter) PrintResult(w io.Writer, n *jnode.Node) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(n)
}
