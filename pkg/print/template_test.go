package print

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/soluble-ai/go-jnode"
)

func TestTemplate(t *testing.T) {
	w := &bytes.Buffer{}
	n := jnode.NewObjectNode()
	n.Put("greeting", "New").Put("subject", "world")
	printer := &TemplatePrinter{
		Template: "Hello {{.greeting}}!. This is a new {{.subject}}",
	}
	printer.PrintResult(w, n)
	fmt.Println(" The result is: " + w.String())
	if s := w.String(); !strings.Contains(s, "New") || !strings.Contains(s, "world") {
		t.Error(s)
	}
}
