package print

import (
	"io"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type TemplatePrinter struct {
	Template string
}

func (tp *TemplatePrinter) PrintResult(w io.Writer, result *jnode.Node) int {
	m := result.Unwrap()
	t := template.New("print-template").Funcs(sprig.TxtFuncMap())
	if _, err := t.Parse(tp.Template); err != nil {
		log.Errorf("Invalid go print template - {danger:%s}", err.Error())
	} else if err := t.Execute(w, m); err != nil {
		log.Warnf("Template failed - {warning:%s}", err.Error())
	}
	return result.Size()
}
