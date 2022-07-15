package print

import (
	"bytes"
	"strings"
	"testing"

	"github.com/soluble-ai/go-jnode"
)

func TestTemplate(t *testing.T) {
	w := &bytes.Buffer{}
	//n := jnode.NewObjectNode()
	// n.Put("greeting", "New").Put("subject", "world")
	// printer := &TemplatePrinter{
	// 	Template: "Hello {{.greeting}}!. This is a new {{.subject}}",
	// }

	n := jnode.NewArrayNode().Append("test").Append("hemanth")
	printer := &TemplatePrinter{
		Template: "Hello {{len .findings}}!.",
	}

	printer.PrintResult(w, n)
	if s := w.String(); !strings.Contains(s, "2") {
		t.Error(s)
	}
}
