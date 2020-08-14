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

import "github.com/soluble-ai/go-jnode"

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
		numeric := false
		if name[0] == '-' {
			name = name[1:]
			asc = false
		}
		if len(name) > 1 && name[0] == '0' {
			numeric = true
			name = name[1:]
		}
		if numeric {
			ni := ei.Path(name).AsFloat()
			nj := ej.Path(name).AsFloat()
			if ni != nj {
				if asc {
					return ni < nj
				}
				return nj < ni
			}
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
