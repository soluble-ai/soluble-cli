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
	r := result
	for _, p := range p.Path {
		r = r.Path(p)
	}
	rows := r.Elements()
	if p.SortBy != nil {
		sort.Sort(&rowsSort{rows, p.SortBy})
	}
	if p.Limit > 0 && len(rows) > p.Limit {
		rows = rows[:p.Limit]
	}
	return rows
}
