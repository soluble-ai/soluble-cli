package print

import (
	"fmt"
	"io"
	"sort"

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
	rows := d.getRows(result)
	sort.Slice(rows, func(i, j int) bool {
		// assume version column is a number
		return rows[i].Path(d.VersionColumn).AsFloat() < rows[j].Path(d.VersionColumn).AsFloat()
	})
	// find diffs
	var previousContent []string
	diffs := make([]string, len(rows))
	previousRow := jnode.MissingNode
	for i, row := range rows {
		content := difflib.SplitLines(row.Path(d.DiffColumn).AsText())
		diff := difflib.UnifiedDiff{
			A:        previousContent,
			B:        content,
			FromFile: fmt.Sprintf("%s=%s", d.VersionColumn, d.Format(d.VersionColumn, previousRow)),
			ToFile:   fmt.Sprintf("%s=%s", d.VersionColumn, d.Format(d.VersionColumn, row)),
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
	width := getLabelWidth(d.LabelColumns)
	for i := len(rows) - 1; i >= v0; i-- {
		if i != len(rows)-1 {
			fmt.Fprintln(w)
		}
		row := rows[i]
		for _, column := range d.LabelColumns {
			printLabel(w, column, width, d.Format(column, row))
		}
		fmt.Fprintln(w, diffs[i])
	}
}
