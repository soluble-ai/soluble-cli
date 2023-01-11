package checkov

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Kustomize struct {
	tools.DirectoryBasedToolOpts
	KustomizeOverlays []string

	overlayPaths []string
}

var _ tools.Interface = (*Kustomize)(nil)

func (k *Kustomize) Name() string {
	return "checkov-kustomize"
}

func (k *Kustomize) Register(cmd *cobra.Command) {
	k.DirectoryBasedToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringSliceVar(&k.KustomizeOverlays, "kustomize-overlays", nil, "Process kustomize overlays in `dirs`.  May be repeated.")
}

func (k *Kustomize) Validate() error {
	if err := k.DirectoryBasedToolOpts.Validate(); err != nil {
		return err
	}
	var dirs []string
	if len(k.KustomizeOverlays) == 0 {
		dirs = []string{k.GetDirectory()}
	} else {
		for _, dir := range k.KustomizeOverlays {
			rel, err := filepath.Rel(k.GetDirectory(), dir)
			if err != nil || strings.HasPrefix(rel, "../") {
				return fmt.Errorf("kustomize overlay %s is not relative to %s", dir, k.GetDirectory())
			}
			dirs = append(dirs, rel)
		}
	}
	for _, dir := range dirs {
		found := false
		for _, name := range []string{"kustomization.yaml", "kustomization.yml"} {
			file := filepath.Join(k.GetDirectory(), dir, name)
			if util.FileExists(file) {
				k.overlayPaths = append(k.overlayPaths, file)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("no kustomization file in %s", filepath.Join(k.GetDirectory(), dir))
		}
	}
	return nil
}

func (k *Kustomize) Run() (*tools.Result, error) {
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
	template.Dir = k.GetDirectory()
	template.Stderr = os.Stderr
	template.Stdout = os.Stderr
	exec := k.ExecuteCommand(template)
	if !exec.ExpectExitCode(0) {
		log.Errorf("{primary:Kustomize template} failed.")
		return exec.ToResult(k.GetDirectory()), nil
	}
	checkov := &Tool{
		DirectoryBasedToolOpts: k.DirectoryBasedToolOpts,
		Framework:              "kubernetes",
		workingDir:             outDirectory,
		pathTranslationFunc: func(s string) string {
			return k.kustomizationName
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
