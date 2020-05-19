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
		n := row.Path(f.name)
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
