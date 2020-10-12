package scanner

import (
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/provider/output"
)

// Output is the runtime engine output
type Output struct {
	ResourceConfig output.AllResourceConfigs
	Violations     policy.EngineOutput
}
