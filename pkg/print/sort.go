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
		if name[0] == '-' {
			name = name[1:]
			asc = false
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
