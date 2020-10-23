package tools

import (
	"github.com/soluble-ai/go-jnode"
)

type Interface interface {
	Run() (*Result, error)
	Name() string
}

type Result struct {
	Data         *jnode.Node
	Values       map[string]string
	PrintPath    []string
	PrintColumns []string
}
