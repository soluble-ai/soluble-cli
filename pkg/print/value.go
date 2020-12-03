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
	Format string
}

var _ Interface = &ValuePrinter{}

var valueSpecRe = regexp.MustCompile(`value\((.*)\)`)

func (p *ValuePrinter) PrintResult(w io.Writer, result *jnode.Node) int {
	m := valueSpecRe.FindStringSubmatch(p.Format)
	if m[1] == "" {
		log.Warnf("invalid value specifier {warning:%s} - must be in the form 'value(name)'", p.Format)
		os.Exit(2)
	}
	name := strings.Split(m[1], ".")
	if p.Path == nil {
		n := Nav(result, name)
		if !n.IsMissing() {
			fmt.Fprintln(w, n.AsText())
			return 1
		}
		return 0
	} else {
		rows := p.GetRows(result)
		for _, row := range rows {
			n := Nav(row, name)
			if !n.IsMissing() {
				fmt.Fprintln(w, n.AsText())
			}
		}
		return len(rows)
	}
}
