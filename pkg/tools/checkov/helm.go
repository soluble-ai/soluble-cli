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

// Despite checkov directly supporting helm it's very buggy.  So instead we'll
// generate the templates ourselves with "helm template" and run checkov on
// the resulting templates much the same way we do for the CDK.
type Helm struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Interface = (*Helm)(nil)

func (h *Helm) Name() string {
	return "checkov-helm"
}

func (h *Helm) Register(cmd *cobra.Command) {
	h.DirectoryBasedToolOpts.Register(cmd)
}

func (h *Helm) Validate() error {
	if err := h.DirectoryBasedToolOpts.Validate(); err != nil {
		return err
	}
	if !util.FileExists(filepath.Join(h.GetDirectory(), "Chart.yaml")) {
		return fmt.Errorf("%s does not contain Chart.yaml", h.GetDirectory())
	}
	return nil
}

func (h *Helm) Run() (*tools.Result, error) {
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
	template.Dir = h.GetDirectory()
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
		result.IACPlatform = "helm"
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
