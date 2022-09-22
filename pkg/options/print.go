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
	"github.com/soluble-ai/soluble-cli/pkg/options/templates"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type PrintOpts struct {
	OutputFormat        []string
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
	Template            []string
	Formatters          map[string]print.Formatter
	ComputedColumns     map[string]print.ColumnFunction
	DiffContextSize     int

	// This is a workaround for the inability of the table printer to
	// accumluate results across an array.  See tools/command.go for how
	// this is used.
	PrintTableDataTransform func(*jnode.Node) *jnode.Node
	outputSource            func() io.Writer
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
			flags.StringSliceVar(&p.Template, "print-template", nil,
				"Print the output with a go `template`.  The template argument may begin with @ in which case the template is read from a file.  If the argument is in the format tmpl=file, then write the output to a file.  May be repeated.")
			flags.StringSliceVar(&p.OutputFormat, "format", nil,
				"Use this output `format` where format is one of: table, yaml, json, none, csv, atlantis, count, or value(name).  If the argument is in the form format=file, then write the output to a file.  May be repeated.")
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

The output format can be selected with the --format flag.  For commands that
support tabular data the default output format is "table"; otherwise the
default is "yaml".

The "value(name)" format prints only the "name" attribute from the results
(from each row if printing tabular data.)

The "count" format prints the number of rows in the result.

Output format "none" suppresses printing.

If the output format is in the format "format=file" then the output
will be written to a file instead of on stdout.  Use "default=file" to
write the default output format to a file.

The --format flag may be given multiple times (or have a comma-separated 
argument), in which case the result is printed once for each argument value.

--print-template flag prints using a Go text template (see
https://pkg.go.dev/text/template for usage.)  The sprig template functions
are available (see http://masterminds.github.io/sprig/).  The input document 
is the JSON result (use "--format json" to examine.)  If the argument
is in the form "template=file" then the result written to a file instead
of printed on stdout.

--print-template may be specified more than once (or its argument may be comma
separated), in which case the result is printed once for each value.

If all --format and --print-template flags are being redirected to a file,
then the default output will still be written to stdout.  To suppress this
output use "--format=none".

For example:

... --format json=output.json \
    --print-template '{{ len . }}'=len.txt

Will write the JSON result to "output.json" and the length of the input 
document to "len.txt", and write the default output to stdout.

The "soluble print" command is useful for experimenting with output formats.

Sorting:

Tabular output can be sorted by one or more columns.  Examples:

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

func getFormatFileOutput(format string) (string, string) {
	eq := strings.IndexRune(format, '=')
	if eq >= 0 {
		return format[0:eq], format[eq+1:]
	}
	return format, ""
}

func (p *PrintOpts) GetPrinter() (print.Interface, error) {
	cp := &chainedPrinter{}
	fileCount := 0
	for _, template := range p.Template {
		template, file := getFormatFileOutput(template)
		tp := &print.TemplatePrinter{Template: template}
		cp.AddPrinter(tp, file)
		if file != "" {
			fileCount++
		}
	}
	for _, format := range p.OutputFormat {
		format, file := getFormatFileOutput(format)
		pr, err := p.getSinglePrinter(format)
		if err != nil {
			return nil, err
		}
		cp.AddPrinter(pr, file)
		if file != "" {
			fileCount++
		}
	}
	if fileCount == len(cp.Printers) {
		// include the default output if everything is going to a file
		pr, err := p.getSinglePrinter("default")
		if err != nil {
			return nil, err
		}
		cp.Printers = append(cp.Printers, pr)
	}
	return cp, nil
}

func (p *PrintOpts) getSinglePrinter(outputFormat string) (print.Interface, error) {
	if outputFormat == "" || outputFormat == "default" {
		outputFormat = p.DefaultOutputFormat
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
		return &tableDataTransformPrinter{
			Printer: &print.CSVPrinter{
				NoHeaders:   p.NoHeaders,
				Columns:     p.getEffectiveColumns(),
				PathSupport: p.getPathSupport(),
				Formatters:  p.Formatters,
			},
			Transform: p.PrintTableDataTransform,
		}, nil
	case "value":
		vp := &tableDataTransformPrinter{
			Printer: &print.ValuePrinter{
				Format:      outputFormat,
				PathSupport: p.getPathSupport(),
			},
			Transform: p.PrintTableDataTransform,
		}
		return vp, nil
	case "atlantis":
		return &print.TemplatePrinter{
			Template: templates.GetEmbeddedTemplate("atlantis.tmpl"),
		}, nil
	case "table":
		if p.Path == nil {
			return nil, fmt.Errorf("this command does not support --format table")
		}
		return &tableDataTransformPrinter{
			Printer: &print.TablePrinter{
				NoHeaders:   p.NoHeaders,
				Columns:     p.getEffectiveColumns(),
				PathSupport: p.getPathSupport(),
				Formatters:  p.Formatters,
			},
			Transform: p.PrintTableDataTransform,
		}, nil
	case "count":
		if p.Path == nil {
			return nil, fmt.Errorf("this command does not support --format count")
		}
		return &tableDataTransformPrinter{
			Printer: &print.CountPrinter{
				PathSupport: p.getPathSupport(),
			},
			Transform: p.PrintTableDataTransform,
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
		return &tableDataTransformPrinter{
			Printer: &print.VerticalPrinter{
				PathSupport: p.getPathSupport(),
				Columns:     p.getEffectiveColumns(),
				Formatters:  p.Formatters,
			},
			Transform: p.PrintTableDataTransform,
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

func (p *PrintOpts) MustPrintStructResult(result interface{}) {
	n, err := print.ToResult(result)
	if err != nil {
		panic(err)
	}
	p.PrintResult(n)
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
