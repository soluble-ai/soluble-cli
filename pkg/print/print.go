package print

import (
	"io"

	"github.com/soluble-ai/go-jnode"
)

type Interface interface {
	PrintResult(w io.Writer, result *jnode.Node)
}

type Formatter func(n *jnode.Node, columnName string) string
