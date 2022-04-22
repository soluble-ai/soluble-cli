package tools

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
	"github.com/spf13/cobra"
)

type UploadOpt struct {
	DefaultUploadEnabled bool
	UploadEnabled        bool
	GitPRBaseRef         string
	UploadErrors         bool
}

func (o *UploadOpt) Register(cmd *cobra.Command) {
	flags := cmd.Flags()
	uploadUsage := "Upload results to lacework"
	if o.DefaultUploadEnabled {
		uploadUsage = fmt.Sprintf("%s.  Use --upload=false to disable.", uploadUsage)
	}
	flags.BoolVar(&o.UploadEnabled, "upload", o.DefaultUploadEnabled, uploadUsage)
	flags.StringVar(&o.GitPRBaseRef, "git-pr-base-ref", "", "Include in the upload a summary of the diffs from `ref` to HEAD.")
	flags.BoolVar(&o.UploadErrors, "upload-errors", false, "Upload tool logs and diagnostics on failures")
}

func (o *UploadOpt) AppendUploadOptions(dir string, options []api.Option) []api.Option {
	if dir != "" && o.GitPRBaseRef != "" {
		diff := o.getPRDIffText(dir)
		if len(diff) > 0 {
			options = append(options,
				xcp.WithFileFromReader("git_pr_diffs", "git-pr-diffs.txt", bytes.NewReader(diff)))
		}
	}
	return options
}

func (o *UploadOpt) getPRDIffText(dir string) []byte {
	buf := &bytes.Buffer{}
	// #nosec G204
	diff := exec.Command("git", "diff", "--name-status", fmt.Sprintf("%s...HEAD", o.GitPRBaseRef))
	fmt.Fprintf(buf, "# %s\n", strings.Join(diff.Args, " "))
	diff.Dir = dir
	diff.Stdout = buf
	if err := diff.Run(); err != nil {
		log.Warnf("Could not determine PR diffs - {warning:%s}", err)
		return nil
	}
	return buf.Bytes()
}
