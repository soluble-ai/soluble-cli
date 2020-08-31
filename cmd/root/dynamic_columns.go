package root

import (
	"sort"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/model"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

func setDynamicColumns(command model.Command, result *jnode.Node) (*jnode.Node, error) {
	opts := command.(*model.OptionsCommand).PrintOpts
	for i := range opts.Columns {
		if opts.Columns[i] == "*" {
			data := print.Nav(result, opts.Path)
			if data.Size() > 0 {
				row := data.Get(0)
				names := []string{}
				colset := util.NewStringSetWithValues(opts.Columns)
				for name := range row.Entries() {
					if !colset.Contains(name) {
						names = append(names, name)
					}
				}
				// This is kinda hokey, but if we don't sort the column
				// order is just hash order.  Instead we should offer the
				// capability to result model from the server response.
				sort.Strings(names)
				columns := []string{}
				if i > 0 {
					columns = append(columns, opts.Columns[0:i-1]...)
				}
				columns = append(columns, names...)
				if i < len(opts.Columns) {
					columns = append(columns, opts.Columns[i+1:]...)
				}
				opts.Columns = columns
			}
			break
		}
	}
	return result, nil
}

func init() {
	model.RegisterAction("dynamic-columns", setDynamicColumns)
}
