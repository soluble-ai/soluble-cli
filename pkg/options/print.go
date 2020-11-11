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
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/exit"
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
	ComputedColumns     map[string]print.ColumnFunction
	DiffContextSize     int
	ExitErrorNotEmtpy   bool
	output              io.Writer
}

var _ Interface = &PrintOpts{}

func (p *PrintOpts) Register(cmd *cobra.Command) {
	flags := cmd.Flags()
	if p.Path == nil {
		flags.StringVar(&p.OutputFormat, "format", p.DefaultOutputFormat, "Use this output format, where format is one of: yaml, json, value or none")
	} else {
		flags.StringVar(&p.OutputFormat, "format", p.DefaultOutputFormat,
			`Use this output format, where format is one of: table,
yaml, json, none, csv, or value(name).  The value(name) form prints
the value of the attribute 'name'.`)
		flags.BoolVar(&p.NoHeaders, "no-headers", false, "Omit headers when printing tables or csv")
		flags.StringVar(&p.Filter, "filter", "",
			`Restrict results to those that match a filter.  The filter
string can be in the form 'attribute=glob-pattern' or
'attribute!=glob-pattern' to search on attributes, or 'attribute=' to
search for rows that contain an attribute, or just 'glob-pattern' to
search all attributes`)
		if p.WideColumns != nil {
			flags.BoolVar(&p.Wide, "wide", false, "Display more columns (table, csv)")
		}
		flags.StringSliceVar(&p.SortBy, "sort-by", p.DefaultSortBy,
			`Sort by these columns (table, csv).  Use -col to indicate
reverse order, 0col to indicate numeric sort, and -0col to indicate
reverse numeric sort.`)
		flags.IntVar(&p.Limit, "print-limit", 0, "Print no more than this number of rows")
		flags.IntVar(&p.DiffContextSize, "diff-context", 3,
			`When printing diffs, the number of lines to print before and
after a a diff.`)
		flags.BoolVar(&p.ExitErrorNotEmtpy, "error-not-empty", false, "Exit with exit code 2 if the results (after filtering) are not empty")
	}
	AddHiddenOptionsGroup(cmd, &HiddenOptionsGroup{
		Use:         "show-print-options",
		Description: "control how results are printed",
		OptionNames: []string{
			"format", "no-headers", "wide", "sort-by", "filter", "print-limit", "diff-context",
			"exit-not-empty",
		},
	})
}

func (p *PrintOpts) GetPrinter() (print.Interface, error) {
	outputFormat := p.OutputFormat
	switch {
	case strings.HasPrefix(p.OutputFormat, "value("):
		outputFormat = "value"
	case p.Path == nil && p.OutputFormat == "":
		// this is the default if the command hasn't specified a Path to the results
		outputFormat = "yaml"
	case p.Path != nil && p.OutputFormat == "":
		// and this is the default if there is a Path
		outputFormat = "table"
	}
	if p.ExitErrorNotEmtpy {
		switch outputFormat {
		case "table", "csv", "value", "vertical":
			// supported
			break
		default:
			return nil, fmt.Errorf("the output format %s cannot be used with --exit-not-empty", outputFormat)
		}
	}
	switch outputFormat {
	case "none":
		return &print.NonePrinter{}, nil
	case "json":
		return &print.JSONPrinter{}, nil
	case "yaml":
		return &print.YAMLPrinter{}, nil
	case "csv":
		if p.Path == nil {
			return nil, fmt.Errorf("this command does not support --format csv")
		}
		p.Wide = true
		return &print.CSVPrinter{
			NoHeaders:   p.NoHeaders,
			Columns:     p.getEffectiveColumns(),
			PathSupport: p.getPathSupport(),
			Formatters:  p.Formatters,
		}, nil
	case "value":
		vp := print.NewValuePrinter(p.OutputFormat, p.Path, p.SortBy)
		vp.Filter = print.NewFilter(p.Filter)
		return vp, nil
	case "table":
		if p.Path == nil {
			return nil, fmt.Errorf("this command does not support --format table")
		}
		return &print.TablePrinter{
			NoHeaders:   p.NoHeaders,
			Columns:     p.getEffectiveColumns(),
			PathSupport: p.getPathSupport(),
			Formatters:  p.Formatters,
		}, nil
	case "diff":
		if p.Path == nil || p.DiffColumn == "" {
			return nil, fmt.Errorf("this command does not support --format diff")
		}
		return &print.DiffPrinter{
			PathSupport:   p.getPathSupport(),
			DiffColumn:    p.DiffColumn,
			VersionColumn: p.VersionColumn,
			LabelColumns:  p.Columns,
			Context:       p.DiffContextSize,
			Formatters:    p.Formatters,
		}, nil
	case "vertical":
		if p.Path == nil {
			return nil, fmt.Errorf("this command does not support --format vertical")
		}
		return &print.VerticalPrinter{
			PathSupport: p.getPathSupport(),
			Columns:     p.getEffectiveColumns(),
			Formatters:  p.Formatters,
		}, nil
	default:
		return nil, fmt.Errorf("this command does not support --format %s", p.OutputFormat)
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
	printer, err := p.GetPrinter()
	if err != nil {
		log.Errorf("Cannot print results: {warning:%s}", err.Error())
		os.Exit(1)
	}
	n := printer.PrintResult(w, result)
	if p.ExitErrorNotEmtpy && n > 0 {
		exit.Message = fmt.Sprintf("Exiting with error code because there are {danger:%d} results", n)
		exit.Code = 2
	}
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

func (p *PrintOpts) SetColumnFunction(columnName string, computer print.ColumnFunction) {
	if p.ComputedColumns == nil {
		p.ComputedColumns = map[string]print.ColumnFunction{}
	}
	p.ComputedColumns[columnName] = computer
}

func (p *PrintOpts) SetContextValues(context map[string]string) {
}
