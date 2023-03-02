package checkov

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Kustomize struct {
	tools.DirectoryBasedToolOpts
	Include  []string
	Parallel int

	overlays []string
}

type kustomizeOverlayRunResult struct {
	result *tools.Result
	err    error
}

var _ tools.Interface = (*Kustomize)(nil)

func (k *Kustomize) Name() string {
	return "checkov-kustomize"
}

func (k *Kustomize) Register(cmd *cobra.Command) {
	k.DirectoryBasedToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringSliceVar(&k.Include, "include", nil, "Look for kustomize overlays in these directory `patterns`.  Each pattern is a gitignore-style string e.g. '**/prod*' would include all overlays in directories that start with 'prod'.  Use --include '**' to run against all overlays.  May be repeated.  Without this flag the scan will only look in the target directory (non-recursive.)")
	flags.IntVar(&k.Parallel, "parallel", 0, "Generate kustomize templates in N threads")
}

func (k *Kustomize) Validate() error {
	if err := k.DirectoryBasedToolOpts.Validate(); err != nil {
		return err
	}
	if len(k.Include) > 0 {
		include := ignore.CompileIgnoreLines(k.Include...)
		m := k.GetInventory()
		k.overlays = nil
		for _, dir := range m.KustomizeDirectories.Values() {
			if include.MatchesPath(dir) {
				overlay := k.findKustomizationFile(dir)
				if overlay != "" {
					k.overlays = append(k.overlays, overlay)
				}
			}
		}
	} else {
		overlay := k.findKustomizationFile("")
		if overlay == "" {
			return fmt.Errorf("no kustomization file in %s", k.GetDirectory())
		}
		k.overlays = []string{overlay}
	}
	return nil
}

func (k *Kustomize) findKustomizationFile(dir string) string {
	for _, name := range []string{"kustomization.yaml", "kustomization.yml"} {
		file := filepath.Join(dir, name)
		if util.FileExists(filepath.Join(k.GetDirectory(), file)) {
			return file
		}
	}
	return ""
}

func (k *Kustomize) Run() (*tools.Result, error) {
	var (
		combinedResult *tools.Result
		failedResult   *tools.Result
		combinedOuput  strings.Builder
	)
	workers := k.Parallel
	if workers <= 1 {
		workers = 1
	}
	overlayCh := make(chan string, workers)
	runResultCh := make(chan *kustomizeOverlayRunResult)
	defer close(overlayCh)
	for i := 0; i < workers; i++ {
		go k.runWorker(overlayCh, runResultCh)
	}
	sent := 0
	received := 0
loop:
	for {
		var runResult *kustomizeOverlayRunResult
		switch {
		case sent < len(k.overlays):
			select {
			case overlayCh <- k.overlays[sent]:
				sent++
			case runResult = <-runResultCh:
				received++
			}
		case received < len(k.overlays):
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
				if len(k.overlays) == 1 {
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
		combinedResult.ExecuteResult.CombinedOutput = &combinedOuput
		return combinedResult, nil
	}
	if failedResult != nil {
		failedResult.ExecuteResult.CombinedOutput = &combinedOuput
		return failedResult, nil
	}
	return nil, fmt.Errorf("no kustomize templates found")
}

func (k *Kustomize) runWorker(overlayCh <-chan string, resultCh chan<- *kustomizeOverlayRunResult) {
	for {
		overlay, more := <-overlayCh
		if !more {
			return
		}
		result, err := k.runOnce(overlay)
		resultCh <- &kustomizeOverlayRunResult{
			result: result,
			err:    err,
		}
	}
}

func (k *Kustomize) runOnce(overlay string) (*tools.Result, error) {
	if err := k.makeKustomizeAvailable(); err != nil {
		return nil, err
	}
	outDirectory, err := os.MkdirTemp(k.GetDirectory(), ".kustomize*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(outDirectory)
	args := []string{"build", "--output", outDirectory}
	template := exec.Command("kustomize", args...)
	template.Dir = filepath.Join(k.GetDirectory(), filepath.Dir(overlay))
	template.Stderr = os.Stderr
	template.Stdout = os.Stderr
	exec := k.ExecuteCommand(template)
	if !exec.ExpectExitCode(0) {
		log.Errorf("{primary:kustomize template} failed.")
		return exec.ToResult(k.GetDirectory()), nil
	}
	checkov := &Tool{
		DirectoryBasedToolOpts: k.DirectoryBasedToolOpts,
		Framework:              "kubernetes",
		workingDir:             outDirectory,
		pathTranslationFunc: func(s string) string {
			return overlay
		},
	}
	if err := checkov.Validate(); err != nil {
		return nil, err
	}
	result, err := checkov.Run()
	if result != nil {
		result.IACPlatform = tools.Kustomize
	}
	return result, err
}

func (k *Kustomize) makeKustomizeAvailable() error {
	c := exec.Command("kustomize", "version")
	if err := c.Run(); err != nil {
		// kustomize does something odd, only the release tags that start with
		// kustomize/ have builds attached, so we can't use the "latest" logic in
		// the download code.  So default a specific version here.
		return installAndAddToPath("github.com/kubernetes-sigs/kustomize", "kustomize",
			"kustomize/v4.5.5")
	}
	return nil
}
