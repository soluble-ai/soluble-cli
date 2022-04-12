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

var _ tools.Consolidated = (*Helm)(nil)

func (h *Helm) Name() string {
	return "checkov-helm"
}

func (h *Helm) RunAll() (tools.Results, error) {
	var (
		results tools.Results
		errs    error
	)
	inventory := h.GetInventory()
	if len(inventory.HelmCharts.Values()) == 0 {
		return nil, fmt.Errorf("no helm charts found under %s", h.GetDirectory())
	}
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
		toolResult, toolErr := tools.RunSingleAssessment(checkov)
		if toolErr != nil {
			// checkov crashes if the chart is malformed, so just
			// accumulate the errors but keep going with other charts
			errs = multierror.Append(errs,
				fmt.Errorf("checkov failed on %s - %w", checkov.Directory, toolErr))
		} else {
			results = append(results, toolResult)
		}
	}
	return results, errs
}
