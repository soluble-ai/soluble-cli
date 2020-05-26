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
	"sort"

	"github.com/soluble-ai/go-jnode"
)

type PathSupport struct {
	Path   []string
	SortBy []string
	Limit  int
}

func (p *PathSupport) getRows(result *jnode.Node) []*jnode.Node {
	r := nav(result, p.Path)
	rows := r.Elements()
	if p.SortBy != nil {
		sort.Sort(&rowsSort{rows, p.SortBy})
	}
	if p.Limit > 0 && len(rows) > p.Limit {
		rows = rows[:p.Limit]
	}
	return rows
}

func nav(n *jnode.Node, path []string) *jnode.Node {
	for _, p := range path {
		n = n.Path(p)
	}
	return n
}
