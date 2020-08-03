package print

import (
	"fmt"
	"io"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/soluble-ai/go-jnode"
)

type DiffPrinter struct {
	PathSupport
	DiffColumn    string
	VersionColumn string
	LabelColumns  []string
	Context       int
	Formatters
}

func (d *DiffPrinter) PrintResult(w io.Writer, result *jnode.Node) {
	d.SortBy = []string{d.VersionColumn}
	rows := d.getRows(result)
	// find diffs
	var previousContent []string
	diffs := make([]string, len(rows))
	previousRow := jnode.MissingNode
	for i, row := range rows {
		content := difflib.SplitLines(row.Path(d.DiffColumn).AsText())
		diff := difflib.UnifiedDiff{
			A:        previousContent,
			B:        content,
			FromFile: fmt.Sprintf("%s=%s", d.VersionColumn, previousRow.Path(d.VersionColumn).AsText()),
			ToFile:   fmt.Sprintf("%s=%s", d.VersionColumn, row.Path(d.VersionColumn).AsText()),
			Context:  d.Context,
		}
		previousContent = content
		previousRow = row
		text, _ := difflib.GetUnifiedDiffString(diff)
		diffs[i] = text
	}
	v0 := 1
	if len(rows) == 1 {
		v0 = 0
	}
	for i := len(rows) - 1; i >= v0; i-- {
		if i != len(rows)-1 {
			fmt.Fprintln(w)
		}
		row := rows[i]
		for _, column := range d.LabelColumns {
			fmt.Fprintf(w, "%s: %s\n", column, d.Format(column, row))
		}
		fmt.Fprintln(w, diffs[i])
	}
}
