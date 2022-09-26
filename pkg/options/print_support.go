package options

import (
	"io"
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
)

type chainedPrinter struct {
	Printers []print.Interface
}

var _ print.Interface = (*chainedPrinter)(nil)

func (cp *chainedPrinter) PrintResult(w io.Writer, result *jnode.Node) int {
	var n int
	for _, p := range cp.Printers {
		n = p.PrintResult(w, result)
	}
	return n
}

func (cp *chainedPrinter) AddPrinter(p print.Interface, file string) {
	if file == "" {
		cp.Printers = append(cp.Printers, p)
	} else {
		cp.Printers = append(cp.Printers, &filePrinter{
			Printer: p,
			File:    file,
		})
	}
}

type filePrinter struct {
	File    string
	Printer print.Interface
}

var _ print.Interface = (*filePrinter)(nil)

func (fp *filePrinter) PrintResult(w io.Writer, result *jnode.Node) int {
	f, err := os.Create(fp.File)
	if err != nil {
		log.Errorf("Cannot write {warning:%s} - {danger:%s}", fp.File, err.Error())
		return 0
	}
	defer f.Close()
	return fp.Printer.PrintResult(f, result)
}

type tableDataTransformPrinter struct {
	Printer   print.Interface
	Transform func(*jnode.Node) *jnode.Node
}

var _ print.Interface = (*tableDataTransformPrinter)(nil)

func (tp *tableDataTransformPrinter) PrintResult(w io.Writer, result *jnode.Node) int {
	if tp.Transform != nil {
		result = tp.Transform(result)
	}
	return tp.Printer.PrintResult(w, result)
}
