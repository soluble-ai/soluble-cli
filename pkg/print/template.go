package print

import (
	"io"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type TemplatePrinter struct {
	Template string
}

func NewTemplatePrinter(template string) (*TemplatePrinter, error) {
	if len(template) > 0 && template[0] == '@' {
		dat, err := os.ReadFile(template[1:])
		if err != nil {
			return nil, err
		}
		template = string(dat)
	}
	return &TemplatePrinter{
		Template: template,
	}, nil
}

func (tp *TemplatePrinter) PrintResult(w io.Writer, result *jnode.Node) int {
	m := result.Unwrap()
	t := template.New("print-template").Funcs(sprig.TxtFuncMap())
	if _, err := t.Parse(tp.Template); err != nil {
		log.Errorf("Invalid go print template - {danger:%s}", err.Error())
		return 0
	}
	if err := t.Execute(w, m); err != nil {
		log.Warnf("Template failed - {warning:%s}", err.Error())
		return 0
	}
	return result.Size()
}
