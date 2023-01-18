package checkov

import "github.com/soluble-ai/go-jnode"

func mergeResults(dest *jnode.Node, source *jnode.Node) {
	combineArrays(dest, source, "results", "passed_checks")
	combineArrays(dest, source, "results", "failed_checks")
	addSummary(dest, source, "passed")
	addSummary(dest, source, "failed")
	addSummary(dest, source, "skipped")
	addSummary(dest, source, "parsing_errors")
	addSummary(dest, source, "resource_count")
}

func combineArrays(dest *jnode.Node, source *jnode.Node, paths ...string) {
	for i := 1; i < len(paths); i++ {
		path := paths[i-1]
		nsource := source.Path(path)
		if !nsource.IsObject() {
			return
		}
		source = nsource
		ndest := dest.Path(path)
		if !ndest.IsObject() {
			ndest = dest.PutObject(path)
		}
		dest = ndest
	}
	last := paths[len(paths)-1]
	darray := dest.Path(last)
	if !darray.IsArray() {
		darray = dest.PutArray(last)
	}
	sarray := source.Path(last)
	if sarray.IsArray() {
		for _, item := range sarray.Elements() {
			darray.Append(item.Unwrap())
		}
	}
}

func addSummary(dest *jnode.Node, source *jnode.Node, field string) {
	ssum := source.Path("summary")
	if !ssum.IsMissing() {
		s := ssum.Path(field)
		if s.IsNumber() {
			dsum := dest.Path("summary")
			if dsum.IsMissing() {
				dsum = dest.PutObject("summary")
			}
			val := dsum.Path(field).AsInt()
			dsum.Put(field, val+s.AsInt())
		}
	}
}
