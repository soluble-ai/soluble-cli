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

package print

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type ValuePrinter struct {
	PathSupport
	Name []string
}

var _ Interface = &ValuePrinter{}

var valueSpecRe = regexp.MustCompile(`value\((.*)\)`)

func NewValuePrinter(format string, path []string, sortBy []string) *ValuePrinter {
	m := valueSpecRe.FindStringSubmatch(format)
	if m[1] == "" {
		log.Warnf("invalid value specifier {warning:%s} - must be in the form 'value(name)'", format)
		os.Exit(2)
	}
	return &ValuePrinter{
		PathSupport: PathSupport{Path: path, SortBy: sortBy},
		Name:        strings.Split(m[1], "."),
	}
}

func (p *ValuePrinter) PrintResult(w io.Writer, result *jnode.Node) {
	if len(p.Path) == 0 {
		n := Nav(result, p.Name)
		if !n.IsMissing() {
			fmt.Fprintln(w, n.AsText())
		}
	} else {
		for _, row := range p.getRows(result) {
			n := Nav(row, p.Name)
			if !n.IsMissing() {
				fmt.Fprintln(w, n.AsText())
			}
		}
	}
}
