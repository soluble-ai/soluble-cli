package inventory

import (
	"os"
	"sort"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/util"
)

func collapseNestedDirs(values *util.StringSet) {
	// collapse nested directories with markers into single dir
	if values.Len() == 0 {
		return
	}
	dirs := values.Values()
	values.Reset()
	sort.Strings(dirs)
	values.Add(dirs[0])
	p := dirs[0]
	for i := 1; i < len(dirs); i++ {
		if p != "." && (!strings.HasPrefix(dirs[i], p) || (len(dirs[i]) > len(p) && dirs[i][len(p)] != os.PathSeparator)) {
			values.Add(dirs[i])
			p = dirs[i]
		}
	}
}
