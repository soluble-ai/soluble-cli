package checkov

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
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
	var errs error
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
			// checkov crashes if the chart is malformed, so just
			// accumulate the errors but keep going with other charts
			errs = multierror.Append(errs,
				fmt.Errorf("checkov failed on %s - %w", checkov.Directory, err))
		} else {
			// Print the result now, but change file_path so that it's
			// relative to the chart directory.  Note that this only
			// affects printing of the result in the CLI, the result
			// has already been uploaded with the correct repo path.
			for _, f := range result.Findings {
				f.FilePath = fmt.Sprintf("%s/%s", chart, f.FilePath)
			}
			checkov.PrintToolResult(result)
		}
	}
	return nil, errs
}
