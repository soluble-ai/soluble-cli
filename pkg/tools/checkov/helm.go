package checkov

import (
	"fmt"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

type Helm struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Interface = (*Helm)(nil)

func (h *Helm) Name() string {
	return "checkov-helm"
}

func (h *Helm) Run() (*tools.Result, error) {
	inventory := h.GetInventory()
	for _, chart := range inventory.HelmCharts.Values() {
		checkov := &Tool{
			DirectoryBasedToolOpts: h.DirectoryBasedToolOpts,
			Framework:              "helm",
		}
		checkov.Directory = filepath.Join(h.GetDirectory(), chart)
		checkov.Tool = checkov
		if err := checkov.Validate(); err != nil {
			return nil, err
		}
		result, err := checkov.RunTool(false)
		if err != nil {
			return nil, err
		}
		// Print the result now, but change file_path so that it's
		// relative to the chart directory.  Note that this only
		// affects printing of the result in the CLI, the result
		// has already been uploaded with the correct repo path.
		for _, f := range result.Findings {
			f.FilePath = fmt.Sprintf("%s/%s", chart, f.FilePath)
		}
		checkov.PrintToolResult(result)
	}
	return nil, nil
}
