package checkov

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

// Despite checkov directly supporting helm it's very buggy.  So instead we'll
// generate the templates ourselves with "helm template" and run checkov on
// the resulting templates much the same way we do for the CDh.
type Helm struct {
	tools.DirectoryBasedToolOpts
	Include  []string
	Parallel int

	charts []string
}

type helmChartRunResult struct {
	result *tools.Result
	err    error
}

var _ tools.Interface = (*Helm)(nil)

func (h *Helm) Name() string {
	return "checkov-helm"
}

func (h *Helm) Register(cmd *cobra.Command) {
	h.DirectoryBasedToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringSliceVar(&h.Include, "include", nil, "Look for helm charts in these directory `patterns`.  Each pattern is a gitignore-style string e.g. '**/prod*' would include all helm charts in directories that start with 'prod'.  Use --include '**' to run against all helm charts.  May be repeated.  Without this flag the scan will only look in the target directory (non-recursive.)")
	flags.IntVar(&h.Parallel, "parallel", 0, "Generate helm chart templates in `N` threads.  If N < 0, then use NumCPU / -N e.g. '--parallel -1' uses NumCPU and '--parallel -2' uses NumCPU / 2")
}

func (h *Helm) Validate() error {
	if err := h.DirectoryBasedToolOpts.Validate(); err != nil {
		return err
	}
	if len(h.Include) > 0 {
		include := ignore.CompileIgnoreLines(h.Include...)
		m := h.GetInventory()
		h.charts = nil
		vals := m.HelmCharts.Values()
		fmt.Println(vals)
		for _, dir := range m.HelmCharts.Values() {
			if include.MatchesPath(dir) {
				chart := h.findChartFile(dir)
				if chart != "" {
					h.charts = append(h.charts, chart)
				}
			}
		}
	} else {
		chart := h.findChartFile("")
		if chart == "" {
			return fmt.Errorf("no Chart file in %s", h.GetDirectory())
		}
		h.charts = []string{chart}
	}
	return nil
}

func (h *Helm) findChartFile(dir string) string {
	for _, name := range []string{"Chart.yaml", "Chart.yml"} {
		file := filepath.Join(dir, name)
		if util.FileExists(filepath.Join(h.GetDirectory(), file)) {
			return file
		}
	}
	return ""
}
func (h *Helm) Run() (*tools.Result, error) {
	var (
		combinedResult *tools.Result
		failedResult   *tools.Result
		combinedOuput  strings.Builder
	)
	workers := h.Parallel
	switch {
	case workers == 0:
		workers = 1
	case workers < 0:
		cpu := runtime.NumCPU()
		workers = cpu / -workers
		if workers == 0 {
			workers = 1
		}
	}
	chartCh := make(chan string, workers)
	runResultCh := make(chan *helmChartRunResult)
	defer close(chartCh)
	for i := 0; i < workers; i++ {
		go h.runWorker(chartCh, runResultCh)
	}
	sent := 0
	received := 0
loop:
	for {
		var runResult *helmChartRunResult
		switch {
		case sent < len(h.charts):
			select {
			case chartCh <- h.charts[sent]:
				sent++
			case runResult = <-runResultCh:
				received++
			}
		case received < len(h.charts):
			runResult = <-runResultCh
			received++
		default:
			break loop
		}
		if runResult != nil {
			result := runResult.result
			err := runResult.err
			if err != nil && result == nil {
				return nil, err
			}
			if result != nil {
				combinedOuput.WriteString(result.ExecuteResult.CombinedOutput.String())
			}
			if err != nil || result.ExecuteResult.FailureType != "" {
				if result != nil {
					if failedResult == nil {
						failedResult = result
					}
				}
				if len(h.charts) == 1 {
					return result, err
				}
				continue
			}
			if combinedResult == nil {
				combinedResult = result
			} else {
				mergeResults(combinedResult.Data, result.Data)
				combinedResult.Findings = append(combinedResult.Findings, result.Findings...)
			}
		}
	}
	if combinedResult != nil {
		combinedResult.Directory = h.GetDirectory()
		combinedResult.ExecuteResult.CombinedOutput = &combinedOuput
		return combinedResult, nil
	}
	if failedResult != nil {
		failedResult.Directory = h.GetDirectory()
		failedResult.ExecuteResult.CombinedOutput = &combinedOuput
		return failedResult, nil
	}
	return nil, fmt.Errorf("no Helm templates found")
}

func (h *Helm) runWorker(chartCh <-chan string, resultCh chan<- *helmChartRunResult) {
	for {
		chart, more := <-chartCh
		if !more {
			return
		}
		result, err := h.runOnce(chart)
		resultCh <- &helmChartRunResult{
			result: result,
			err:    err,
		}
	}
}

func (h *Helm) runOnce(chart string) (*tools.Result, error) {
	outDirectory, err := os.MkdirTemp(h.GetDirectory(), ".helm*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(outDirectory)
	if err := h.makeHelmAvailable(); err != nil {
		return nil, err
	}
	args := []string{"template", "--dependency-update", "--output-dir", outDirectory, "."}
	template := exec.Command("helm", args...)
	template.Dir = filepath.Join(h.GetDirectory(), filepath.Dir(chart))
	template.Stderr = os.Stderr
	exec := h.ExecuteCommand(template)
	if !exec.ExpectExitCode(0) {
		log.Errorf("{primary:helm template} failed.")
		return exec.ToResult(h.GetDirectory()), nil
	}
	checkov := &Tool{
		DirectoryBasedToolOpts: h.DirectoryBasedToolOpts,
		Framework:              "kubernetes",
		workingDir:             outDirectory,
		pathTranslationFunc: func(s string) string {
			// helm template writes to <out-dir>/<chart-name>/...
			// and checkov reports it as /<chart-name>/...
			// we want to turn that into <dir>/...
			if len(s) > 2 {
				slash := strings.IndexRune(s[1:], '/')
				s = s[slash+2:]
			}
			return s
		},
	}
	if err := checkov.Validate(); err != nil {
		return nil, err
	}
	result, err := checkov.Run()
	if result != nil {
		result.IACPlatform = tools.Helm
	}
	return result, err
}

func (h *Helm) makeHelmAvailable() error {
	c := exec.Command("helm", "version")
	if err := c.Run(); err != nil {
		return installAndAddToPath("github.com/helm/helm", "helm", "")
	}
	return nil
}
