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

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type PrintOpts struct {
	OutputFormat  string
	NoHeaders     bool
	Wide          bool
	Path          []string
	Columns       []string
	WideColumns   []string
	SortBy        []string
	DefaultSortBy []string
	Formatters    map[string]print.Formatter
	output        io.Writer
}

var _ Interface = &PrintOpts{}

func (p *PrintOpts) Register(cmd *cobra.Command) {
	if p.Path != nil {
		cmd.Flags().StringVar(&p.OutputFormat, "format", "", "Use this output format, where format is one of: table, yaml, json, csv")
		cmd.Flags().BoolVar(&p.NoHeaders, "no-headers", false, "Omit headers when printing tables or csv")
		if p.WideColumns != nil {
			cmd.Flags().BoolVar(&p.Wide, "wide", false, "Display more columns (table, csv)")
		}
		cmd.Flags().StringSliceVar(&p.SortBy, "sort-by", p.DefaultSortBy, "Sort by these columns (table, csv)")
	}
}

func (p *PrintOpts) GetPrinter() print.Interface {
	switch p.OutputFormat {
	case "json":
		return &print.JSONPrinter{}
	case "yaml":
		return &print.YAMLPrinter{}
	case "csv":
		if len(p.Path) == 0 {
			log.Errorf("This command does not support the {danger:csv} format")
			os.Exit(2)
		}
		return &print.CSVPrinter{
			NoHeaders:  p.NoHeaders,
			Columns:    p.getEffectiveColumns(),
			Path:       p.Path,
			SortBy:     p.SortBy,
			Formatters: p.Formatters,
		}
	case "table":
		fallthrough
	default:
		if len(p.Path) == 0 {
			if p.OutputFormat == "table" {
				log.Errorf("This command does not support the {danger:table} format")
				os.Exit(2)
			}
			return &print.YAMLPrinter{}
		}
		return &print.TablePrinter{
			NoHeaders:  p.NoHeaders,
			Columns:    p.getEffectiveColumns(),
			Path:       p.Path,
			SortBy:     p.SortBy,
			Formatters: p.Formatters,
		}
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

func (p *PrintOpts) SetFormatter(columnName string, formatter print.Formatter) {
	if p.Formatters == nil {
		p.Formatters = map[string]print.Formatter{}
	}
	p.Formatters[columnName] = formatter
}

func (p *PrintOpts) SetContextValues(context map[string]string) {
}
