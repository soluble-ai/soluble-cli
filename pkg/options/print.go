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
	"io"
	"os"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type PrintOpts struct {
	OutputFormat        string
	DefaultOutputFormat string
	NoHeaders           bool
	Wide                bool
	Path                []string
	Columns             []string
	WideColumns         []string
	DiffColumn          string
	VersionColumn       string
	SortBy              []string
	DefaultSortBy       []string
	Limit               int
	Filter              string
	Formatters          map[string]print.Formatter
	ComputedColumns     map[string]print.ColumnComputer
	DiffContextSize     int
	output              io.Writer
}

var _ Interface = &PrintOpts{}

func (p *PrintOpts) Register(cmd *cobra.Command) {
	if p.Path == nil {
		cmd.Flags().StringVar(&p.OutputFormat, "format", p.DefaultOutputFormat, "Use this output format, where format is one of: yaml, json, value or none")
	} else {
		cmd.Flags().StringVar(&p.OutputFormat, "format", p.DefaultOutputFormat,
			`Use this output format, where format is one of: table,
yaml, json, none, csv, or value(name).  The value(name) form prints
the value of the attribute 'name'.`)
		cmd.Flags().BoolVar(&p.NoHeaders, "no-headers", false, "Omit headers when printing tables or csv")
		cmd.Flags().StringVar(&p.Filter, "filter", "",
			`Restrict results to those that match a filter.  The filter
string can be in the form 'attribute=glob-pattern' or
'attribute!=glob-pattern' to search on attributes, or 'attribute=' to
search for rows that contain an attribute, or just 'glob-pattern' to
search all attributes`)
		if p.WideColumns != nil {
			cmd.Flags().BoolVar(&p.Wide, "wide", false, "Display more columns (table, csv)")
		}
		cmd.Flags().StringSliceVar(&p.SortBy, "sort-by", p.DefaultSortBy,
			`Sort by these columns (table, csv).  Use -col to indicate
reverse order, 0col to indicate numeric sort, and -0col to indicate
reverse numeric sort.`)
		cmd.Flags().IntVar(&p.Limit, "print-limit", 0, "Print no more than this number of rows")
		cmd.Flags().IntVar(&p.DiffContextSize, "diff-context", 3,
			`When printing diffs, the number of lines to print before and
after a a diff.`)
	}
}

func (p *PrintOpts) GetPrinter() print.Interface {
	switch {
	case p.OutputFormat == "none":
		return &print.NonePrinter{}
	case p.OutputFormat == "json":
		return &print.JSONPrinter{}
	case p.Path == nil && (p.OutputFormat == "" || p.OutputFormat == "yaml"):
		return &print.YAMLPrinter{}
	case p.OutputFormat == "csv":
		if p.Path == nil {
			log.Errorf("This command does not support the {danger:csv} format")
			os.Exit(2)
		}
		p.Wide = true
		return &print.CSVPrinter{
			NoHeaders:   p.NoHeaders,
			Columns:     p.getEffectiveColumns(),
			PathSupport: p.getPathSupport(),
			Formatters:  p.Formatters,
		}
	case strings.HasPrefix(p.OutputFormat, "value("):
		vp := print.NewValuePrinter(p.OutputFormat, p.Path, p.SortBy)
		vp.Filter = print.NewFilter(p.Filter)
		return vp
	case p.Path != nil && (p.OutputFormat == "" || p.OutputFormat == "table"):
		return &print.TablePrinter{
			NoHeaders:   p.NoHeaders,
			Columns:     p.getEffectiveColumns(),
			PathSupport: p.getPathSupport(),
			Formatters:  p.Formatters,
		}
	case p.Path != nil && p.OutputFormat == "diff":
		if p.DiffColumn == "" {
			log.Errorf("This command does not support diff output")
			os.Exit(2)
		}
		return &print.DiffPrinter{
			PathSupport:   p.getPathSupport(),
			DiffColumn:    p.DiffColumn,
			VersionColumn: p.VersionColumn,
			LabelColumns:  p.Columns,
			Context:       p.DiffContextSize,
			Formatters:    p.Formatters,
		}
	case p.Path != nil && p.OutputFormat == "vertical":
		return &print.VerticalPrinter{
			PathSupport: p.getPathSupport(),
			Columns:     p.getEffectiveColumns(),
			Formatters:  p.Formatters,
		}
	default:
		log.Errorf("This command does not support the {danger:%s} format", p.OutputFormat)
		os.Exit(2)
		return nil
	}
}

func (p *PrintOpts) getPathSupport() print.PathSupport {
	return print.PathSupport{
		Filter:          print.NewFilter(p.Filter),
		Path:            p.Path,
		SortBy:          p.SortBy,
		ComputedColumns: p.ComputedColumns,
	}
}

func (p *PrintOpts) PrintResult(result *jnode.Node) {
	var w io.Writer
	if p.output != nil {
		w = p.output
	} else {
		w = os.Stdout
	}
	p.GetPrinter().PrintResult(w, result)
}

// Returns all the columns that should be included in the result,
// in order.  If Wide is set, then union(p.Columns, p.WideColumns).
// Otherwise all p.Columns not in p.WideColumns.
func (p *PrintOpts) getEffectiveColumns() []string {
	if p.Wide {
		result := util.NewStringSet()
		for _, c := range p.Columns {
			result.Add(c)
		}
		for _, wc := range p.WideColumns {
			result.Add(wc)
		}
		return result.Values()
	}
	wc := util.NewStringSet().AddAll(p.WideColumns...)
	columns := make([]string, 0, len(p.Columns))
	for _, c := range p.Columns {
		if !wc.Contains(c) {
			columns = append(columns, c)
		}
	}
	return columns
}

func (p *PrintOpts) SetFormatter(columnName string, formatter print.Formatter) {
	if p.Formatters == nil {
		p.Formatters = map[string]print.Formatter{}
	}
	p.Formatters[columnName] = formatter
}

func (p *PrintOpts) SetColumnComputer(columnName string, computer print.ColumnComputer) {
	if p.ComputedColumns == nil {
		p.ComputedColumns = map[string]print.ColumnComputer{}
	}
	p.ComputedColumns[columnName] = computer
}

func (p *PrintOpts) SetContextValues(context map[string]string) {
}
