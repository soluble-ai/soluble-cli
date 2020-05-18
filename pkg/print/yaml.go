package print

import (
	"fmt"
	"io"

	"github.com/soluble-ai/go-jnode"
	"gopkg.in/yaml.v2"
)

type YAMLPrinter struct{}

var _ Interface = &YAMLPrinter{}

func (p *YAMLPrinter) PrintResult(w io.Writer, n *jnode.Node) {
	s, _ := yaml.Marshal(n.Unwrap())
	fmt.Fprintln(w, string(s))
}
