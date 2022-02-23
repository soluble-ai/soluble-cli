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
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	Filter              []string
	Template            string
	Formatters          map[string]print.Formatter
	ComputedColumns     map[string]print.ColumnFunction
	DiffContextSize     int
	outputSource        func() io.Writer
}

var _ Interface = &PrintOpts{}

func GetPrintOptionsGroupHelpCommand() *cobra.Command {
	opts := &PrintOpts{}
	return opts.GetPrintOptionsGroup().GetHelpCommand()
}

func (p *PrintOpts) GetPrintOptionsGroup() *HiddenOptionsGroup {
	return &HiddenOptionsGroup{
		Name: "print-options",
		Long: "These options control how results are printed",
		CreateFlagsFunc: func(flags *pflag.FlagSet) {
			flags.StringVar(&p.Template, "print-template", "",
				"The go `template` to print with.  If the template begins with @, then read the template from a file.")
			flags.StringVar(&p.OutputFormat, "format", "",
				"Use this output `format` where format is one of: table, yaml, json, none, csv, template, count, or value(name).")
			flags.BoolVar(&p.NoHeaders, "no-headers", false, "Omit headers when printing tables or csv")
			flags.StringSliceVar(&p.Filter, "filter", nil, "Print results that match a `filter`.  May be repeated.")
			flags.BoolVar(&p.Wide, "wide", false, "Display more columns (table, csv)")
			flags.StringSliceVar(&p.SortBy, "sort-by", p.DefaultSortBy,
				"Sort by these `columns`")
			flags.IntVar(&p.Limit, "print-limit", 0, "Print no more than this `number` of rows")
			flags.IntVar(&p.DiffContextSize, "diff-context", 3,
				"When printing diffs, the number of `lines` to print before and after a a diff.")
		},
		Example: `
Output formats:

The output format can be selected with the --format flag.  If --print-template is given
the default format in "template".  For commands that support tabular data the default
output format is "table"; otherwise the default is "yaml".

The "value(name)" format prints only the "name" attribute from the results (from each row
if printing tabular data.)

The "count" format prints the number of rows in the result.

Sorting:

The tabular output can be sorted by one or more columns.  Examples:

 ... --sort-by col1,-col2    ;# ascending col1, descending col2, lexigraphical
 ... --sort-by 0col1         ;# ascending, numerical
 ... --sort-by -0col1        ;# descending numerical

Filtering:

Tabular output can be filtered with glob-style patterns.  Examples:

 ... --filter col1=pattern   ;# print rows that match on col1
 ... --filter col1!=pattern  ;# print rows that don't match
 ... --filter pattern        ;# print rows that match on any column`,
	}
}

func (p *PrintOpts) Register(cmd *cobra.Command) {
	p.GetPrintOptionsGroup().Register(cmd)
	p.outputSource = cmd.OutOrStdout
}

func (p *PrintOpts) GetPrinter() (print.Interface, error) {
	outputFormat := p.OutputFormat
	if outputFormat == "" {
		if p.Template != "" {
			outputFormat = "template"
		} else {
			outputFormat = p.DefaultOutputFormat
		}
	}
	outputFormatType := outputFormat
	switch {
	case strings.HasPrefix(outputFormat, "value("):
		outputFormatType = "value"
	case p.Path == nil && outputFormat == "":
		// this is the default if the command hasn't specified a Path to the results
		outputFormatType = "yaml"
	case p.Path != nil && outputFormat == "":
		// and this is the default if there is a Path
		outputFormatType = "table"
	}
	switch outputFormatType {
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
		vp := &print.ValuePrinter{
			Format:      outputFormat,
			PathSupport: p.getPathSupport(),
		}
		return vp, nil
	case "template":
		if p.Template == "" {
			return nil, fmt.Errorf("--print-template must be specified for --format template")
		}
		template := p.Template
		if template[0] == '@' {
			dat, err := os.ReadFile(template[1:])
			if err != nil {
				return nil, err
			}
			template = string(dat)
		}
		return &print.TemplatePrinter{
			Template: template,
		}, nil
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
	case "count":
		if p.Path == nil {
			return nil, fmt.Errorf("this command does not support --format count")
		}
		return &print.CountPrinter{
			PathSupport: p.getPathSupport(),
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
		Filter:          print.NewAndFilter(p.Filter),
		Path:            p.Path,
		SortBy:          p.SortBy,
		ComputedColumns: p.ComputedColumns,
		Limit:           p.Limit,
	}
}

func (p *PrintOpts) PrintResult(result *jnode.Node) {
	var w io.Writer
	if p.outputSource != nil {
		w = p.outputSource()
	} else {
		w = os.Stdout
	}
	printer, err := p.GetPrinter()
	if err != nil {
		log.Errorf("Cannot print results: {warning:%s}", err.Error())
		os.Exit(1)
	}
	_ = printer.PrintResult(w, result)
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
