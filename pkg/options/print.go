// Copyright 2020 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package options

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Formatter func(n *jnode.Node, columnName string) string

type HasPrintOpts interface {
	GetPrintOpts() *PrintOpts
}

type PrintOpts struct {
	Output        io.Writer
	Full          bool
	Wide          bool
	Path          []string
	Columns       []string
	WideColumns   []string
	SortBy        []string
	DefaultSortBy []string
	Formatters    map[string]Formatter
	Transformer   func(*jnode.Node) *jnode.Node
	w             *tabwriter.Writer
}

var _ Interface = &PrintOpts{}

type rowsSort struct {
	values []*jnode.Node
	sortBy []string
}

func (r *rowsSort) Len() int {
	return len(r.values)
}
func (r *rowsSort) Swap(i, j int) {
	r.values[i], r.values[j] = r.values[j], r.values[i]
}
func (r *rowsSort) Less(i, j int) bool {
	ei := r.values[i]
	ej := r.values[j]
	for _, name := range r.sortBy {
		asc := true
		if name[0] == '-' {
			name = name[1:]
			asc = false
		}
		vi := ei.Path(name).AsText()
		vj := ej.Path(name).AsText()
		if vi != vj {
			if asc {
				return vi < vj
			}
			return vj < vi
		}
	}
	return false
}

func (p *PrintOpts) GetPrintOpts() *PrintOpts {
	return p
}

func (p *PrintOpts) Register(cmd *cobra.Command) {
	if p.Path != nil {
		cmd.Flags().BoolVar(&p.Full, "full", false, "Disable full details in YAML format")
		if p.WideColumns != nil {
			cmd.Flags().BoolVar(&p.Wide, "wide", false, "Display more columns")
		}
		cmd.Flags().StringSliceVar(&p.SortBy, "sort-by", p.DefaultSortBy, "Sort by these columns")
	}
}

func (p *PrintOpts) AddTransformer(t func(*jnode.Node) *jnode.Node) {
	next := p.Transformer
	p.Transformer = func(n *jnode.Node) *jnode.Node {
		n = t(n)
		if next != nil {
			n = next(n)
		}
		return n
	}
}

func (p *PrintOpts) isFull() bool {
	return p.Full || len(p.Path) == 0
}

func (p *PrintOpts) PrintResult(result *jnode.Node) {
	if p.isFull() {
		p.PrintResultYAML(result)
	} else {
		p.PrintHeader()
		p.PrintRows(result)
		p.Flush()
	}
}

func (p *PrintOpts) PrintRows(result *jnode.Node) {
	if p.isFull() {
		p.PrintResultYAML(result)
		return
	}
	r := result
	for _, p := range p.Path {
		r = r.Path(p)
	}
	rows := []*jnode.Node{}
	for _, e := range r.Elements() {
		if p.Transformer != nil {
			e = p.Transformer(e)
		}
		rows = append(rows, e)
	}
	if p.SortBy != nil {
		sort.Sort(&rowsSort{rows, p.SortBy})
	}
	for _, row := range rows {
		for i, c := range p.getEffectiveColumns() {
			if i > 0 {
				fmt.Fprint(p.w, "\t")
			}
			fmt.Fprint(p.w, p.Format(c, row))
		}
		fmt.Fprint(p.w, "\n")
	}
}

func (p *PrintOpts) getEffectiveColumns() []string {
	if p.Wide {
		return p.Columns
	}
	columns := make([]string, 0, len(p.Columns))
	for _, c := range p.Columns {
		if !util.StringSliceContains(p.WideColumns, c) {
			columns = append(columns, c)
		}
	}
	return columns
}

func (p *PrintOpts) PrintHeader() {
	if p.isFull() {
		return
	}
	p.w = tabwriter.NewWriter(p.GetOutput(), 5, 0, 1, ' ', 0)
	for i, c := range p.getEffectiveColumns() {
		if i > 0 {
			fmt.Fprint(p.w, "\t")
		}
		fmt.Fprint(p.w, toHeader(c))
	}
	fmt.Fprint(p.w, "\n")
}

func (p *PrintOpts) Flush() {
	if p.w != nil {
		p.w.Flush()
		p.w = nil
	}
}

func (p *PrintOpts) Format(columnName string, n *jnode.Node) string {
	if p.Formatters != nil {
		formatter := p.Formatters[columnName]
		if formatter != nil {
			return formatter(n, columnName)
		}
	}
	var s string
	switch {
	case strings.HasSuffix(columnName, "Ts+"):
		return RelativeTimestampFormatter(n, columnName)
	case strings.HasSuffix(columnName, "Ts"):
		return TimestampFormatter(n, columnName)
	default:
		s = n.Path(columnName).AsText()
	}
	return s
}

func (p *PrintOpts) SetFormatter(columnName string, formatter Formatter) {
	if p.Formatters == nil {
		p.Formatters = map[string]Formatter{}
	}
	p.Formatters[columnName] = formatter
}

func (p *PrintOpts) GetOutput() io.Writer {
	if p.Output != nil {
		return p.Output
	}
	return os.Stdout
}

func (p *PrintOpts) PrintResultYAML(result interface{}) {
	output := p.GetOutput()
	if value, ok := result.(*jnode.Node); ok {
		result = value.Unwrap()
	}
	s, _ := yaml.Marshal(result)
	fmt.Fprintln(output, string(s))
}

func toHeader(c string) string {
	w := &bytes.Buffer{}
	var wasUpper int
	if c[len(c)-1] == '+' {
		c = c[0 : len(c)-1]
	}
	for i, ch := range c {
		upper := unicode.IsUpper(ch)
		if i > 0 && wasUpper == 0 && upper {
			w.WriteRune('-')
		}
		if ch == '_' {
			w.WriteRune('-')
			wasUpper = -1
		} else {
			w.WriteRune(unicode.ToUpper(ch))
			wasUpper = 0
			if upper {
				wasUpper = 1
			}
		}
	}
	return w.String()
}

func (p *PrintOpts) SetContextValues(context map[string]string) {
}
