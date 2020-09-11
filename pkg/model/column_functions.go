package model

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/print"
)

type ColumnFunctionType string

type ColumnFunctionMaker func(name string, args []string) (print.ColumnFunction, error)

var (
	columnFunctions = map[string]ColumnFunctionMaker{}
	namePattern     = regexp.MustCompile(`([^(]+)(\(([^)]+)\))?`)
)

func (t ColumnFunctionType) validate() error {
	name, _ := t.parse()
	if _, ok := columnFunctions[name]; !ok {
		return fmt.Errorf("invalid column function %s", t)
	}
	if _, err := t.GetColumnFunction(); err != nil {
		return err
	}
	return nil
}

func (t ColumnFunctionType) parse() (name string, args []string) {
	name = string(t)
	args = nil
	m := namePattern.FindStringSubmatch(name)
	if m != nil {
		name = m[1]
		if m[3] == "" {
			args = []string{}
		} else {
			args = strings.Split(m[3], ",")
		}
	}
	return
}

func (t ColumnFunctionType) GetColumnFunction() (print.ColumnFunction, error) {
	name, args := t.parse()
	return columnFunctions[name](name, args)
}

func RegisterColumnFunction(name string, f print.ColumnFunction) {
	columnFunctions[name] = func(name string, args []string) (print.ColumnFunction, error) {
		return f, nil
	}
}

func RegisterParameterizedColumnFunction(name string, m ColumnFunctionMaker) {
	columnFunctions[name] = m
}

func emptyIfNaN(x float64) interface{} {
	if math.IsNaN(x) {
		return ""
	}
	return x
}

func minMaxColumnFunctions(name string, args []string) (print.ColumnFunction, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%s function requires 1 argument", name)
	}
	return func(row *jnode.Node) interface{} {
		values := row.Path(args[0])
		if values.Size() == 0 {
			return ""
		}
		min := math.NaN()
		max := math.NaN()
		for _, e := range values.Elements() {
			if !e.IsMissing() {
				if math.IsNaN(min) {
					min = e.AsFloat()
					max = e.AsFloat()
				}
				min = math.Min(min, e.AsFloat())
				max = math.Max(max, e.AsFloat())
			}
		}
		switch name {
		case "min":
			return emptyIfNaN(min)
		case "max":
			return emptyIfNaN(max)
		case "range":
			fallthrough
		default:
			if math.IsNaN(min) {
				return ""
			}
			if min == max {
				return min
			}
			return fmt.Sprintf("%d - %d", int64(min), int64(max))
		}
	}, nil
}

func init() {
	RegisterParameterizedColumnFunction("min", minMaxColumnFunctions)
	RegisterParameterizedColumnFunction("max", minMaxColumnFunctions)
	RegisterParameterizedColumnFunction("range", minMaxColumnFunctions)
}
