package print

import (
	"sort"
	"testing"

	"github.com/soluble-ai/go-jnode"
)

func TestSortBy(t *testing.T) {
	rows := []*jnode.Node{
		jnode.NewObjectNode().Put("one", "b").Put("two", "b"),
		jnode.NewObjectNode().Put("one", "b").Put("two", "a"),
		jnode.NewObjectNode().Put("one", "a").Put("two", "a"),
		jnode.NewObjectNode().Put("one", "a").Put("two", "b"),
		jnode.NewObjectNode().Put("one", "a").Put("two", "a").Put("three", "z"),
	}
	sort.Sort(&rowsSort{rows, []string{"one", "two", "three"}})
	assertEqual(t, rows[0], "a", "a", "")
	assertEqual(t, rows[1], "a", "a", "z")
	assertEqual(t, rows[2], "a", "b", "")
	assertEqual(t, rows[3], "b", "a", "")
	assertEqual(t, rows[4], "b", "b", "")
}

func assertEqual(t *testing.T, n *jnode.Node, one, two, three string) {
	if n.Path("one").AsText() != one || n.Path("two").AsText() != two ||
		n.Path("three").AsText() != three {
		t.Error(n)
	}
}
