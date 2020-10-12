package policy

import (
	"encoding/xml"

	"github.com/soluble-ai/soluble-cli/pkg/provider/output"
	"github.com/soluble-ai/soluble-cli/pkg/results"
)

// EngineInput Contains data used as input to the engine
type EngineInput struct {
	InputData *output.AllResourceConfigs
}

// EngineOutput Contains data output from the engine
type EngineOutput struct {
	XMLName                 xml.Name `json:"-" yaml:"-" xml:"results"`
	*results.ViolationStore `json:"results" yaml:"results" xml:"results"`
}
