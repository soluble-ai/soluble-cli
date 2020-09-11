package model

import "testing"

var columnFunctionParseTestCases = []struct {
	s    string
	name string
	args []string
}{
	{"max(foo)", "max", []string{"foo"}},
	{"min", "min", []string{}},
	{"cons(first,second)", "cons", []string{"first", "second"}},
}

func TestParseColumnFunction(t *testing.T) {
	for _, c := range columnFunctionParseTestCases {
		name, args := ColumnFunctionType(c.s).parse()
		if name != c.name {
			t.Error(name, c.name)
		}
		if len(args) != len(c.args) {
			t.Error(len(args), len(c.args))
		} else {
			for i := range args {
				if args[i] != c.args[i] {
					t.Error(args, c.args)
				}
			}
		}
	}
}
