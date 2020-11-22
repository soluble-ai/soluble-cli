package tools

import "github.com/soluble-ai/soluble-cli/pkg/options"

type Interface interface {
	options.Interface
	GetToolOptions() *ToolOpts
	Run() (*Result, error)
	Name() string
}
