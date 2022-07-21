package print

import (
	"bytes"
	"testing"

	"github.com/soluble-ai/go-jnode"
)

func TestTemplate(t *testing.T) {
	w := &bytes.Buffer{}
	n := jnode.NewArrayNode().Append("test").Append("hemanth")
	printer := &TemplatePrinter{
		Template: "Hello {{len .}}!.",
	}
	printer.PrintResult(w, n)
	if s := w.String(); s != "Hello 2!." {
		t.Error(s)
	}
}
