package tools

import (
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type Interface interface {
	Run() (*Result, error)
	Name() string
}

type Result struct {
	Data         *jnode.Node
	Values       map[string]string
	Directory    string
	Files        *util.StringSet
	PrintPath    []string
	PrintColumns []string
}

func (r *Result) AddFile(path string) *Result {
	if r.Files == nil {
		r.Files = util.NewStringSet()
	}
	r.Files.Add(path)
	return r
}

func (r *Result) AddValue(name, value string) *Result {
	if r.Values == nil {
		r.Values = map[string]string{}
	}
	r.Values[name] = value
	return r
}
