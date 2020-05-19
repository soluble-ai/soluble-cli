package print

import (
	"testing"

	"github.com/soluble-ai/go-jnode"
)

func TestFilter(t *testing.T) {
	row := jnode.NewObjectNode().Put("name", "value").Put("greeting", "hello")
	if n := NewFilter("hello"); n.name != "" || n.g == nil || !n.matches(row) {
		t.Error(n)
	}
	if n := NewFilter("name="); n.name != "name" || !n.matches(row) {
		t.Error(n)
	}
	if n := NewFilter(""); !n.matches(row) {
		t.Error(n)
	}
	if n := NewFilter("world"); n.matches(row) {
		t.Error(n)
	}
	if n := NewFilter("name=joe*"); n.matches(row) {
		t.Error(n)
	}
	if n := NewFilter("name=v*"); !n.matches(row) {
		t.Error(n)
	}
	if n := NewFilter("name!=value"); n.matches(row) {
		t.Error(n)
	}
}
