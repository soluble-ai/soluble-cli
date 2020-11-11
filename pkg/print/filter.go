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
	"regexp"
	"strings"

	"github.com/gobwas/glob"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type Filter struct {
	name string
	g    glob.Glob
	not  bool
}

var nvRegexp = regexp.MustCompile("([^!=]+!?=)?(.*)")

func NewFilter(s string) Filter {
	m := nvRegexp.FindStringSubmatch(s)
	f := Filter{}
	if m[1] != "" {
		f.name = m[1]
		if strings.HasSuffix(f.name, "!=") {
			f.not = true
			f.name = f.name[:len(f.name)-2]
		} else {
			f.name = f.name[:len(f.name)-1]
		}
	}
	pat := m[2]
	if pat != "" {
		var err error
		f.g, err = glob.Compile(pat)
		if err != nil {
			log.Warnf("Ignoring invalid filter {info:%s} - {danger:%s}", pat, err.Error())
			return Filter{}
		}
	}
	return f
}

func (f Filter) matches(row *jnode.Node) bool {
	if f.name != "" {
		// should fix - print columns supports paths a.b but filtering does not
		n := row.Path(f.name)
		if n.IsMissing() {
			// slow path - try case insensitive key name
			for k, v := range row.Entries() {
				if strings.EqualFold(k, f.name) {
					n = v
					break
				}
			}
		}
		if n.IsMissing() {
			return f.not
		}
		if f.g != nil && !f.g.Match(n.AsText()) {
			return f.not
		}
		return !f.not
	}
	if f.g != nil {
		for _, e := range row.Entries() {
			if f.g.Match(e.AsText()) {
				return !f.not
			}
		}
		return f.not
	}
	return !f.not
}
