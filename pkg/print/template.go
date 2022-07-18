package print

import (
	"io"
	"text/template"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type TemplatePrinter struct {
	Template string
}

func (tp *TemplatePrinter) PrintResult(w io.Writer, result *jnode.Node) int {
	if result.IsArray() {
		n := jnode.NewObjectNode()
		n.Put("findings", result)
		result = n
	}
	m := result.ToMap()
	t := template.New("print-template")
	if _, err := t.Parse(tp.Template); err != nil {
		log.Errorf("Invalid go print template - {danger:%s}", err.Error())
	} else if err := t.Execute(w, m); err != nil {
		log.Warnf("Template failed - {warning:%s}", err.Error())
	}
	return result.Size()
}
