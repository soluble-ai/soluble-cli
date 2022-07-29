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

type UploadOpts struct {
	DefaultUploadEnabled bool
	UploadEnabled        bool
	GitPRBaseRef         string
	UploadErrors         bool
	CompressResults      bool
}

func (o *UploadOpts) Register(cmd *cobra.Command) {
	flags := cmd.Flags()
	uploadUsage := "Upload results to lacework"
	if o.DefaultUploadEnabled {
		uploadUsage = fmt.Sprintf("%s.  Use --upload=false to disable.", uploadUsage)
	}
	flags.BoolVar(&o.UploadEnabled, "upload", o.DefaultUploadEnabled, uploadUsage)
	flags.Lookup("upload").Hidden = true
	flags.StringVar(&o.GitPRBaseRef, "git-pr-base-ref", "", "Include in the upload a summary of the diffs from `ref` to HEAD.")
	flags.BoolVar(&o.UploadErrors, "upload-errors", false, "Upload tool logs and diagnostics on failures")
	flags.BoolVar(&o.CompressResults, "x-compress-results", false, "Compress results before uploading.")
	flags.Lookup("x-compress-results").Hidden = true
}

func (o *UploadOpts) AppendUploadOptions(dir string, options []api.Option) []api.Option {
	if dir != "" && o.GitPRBaseRef != "" {
		diff := o.getPRDIffText(dir)
		if len(diff) > 0 {
			options = append(options,
				xcp.WithFileFromReader("git_pr_diffs", "git-pr-diffs-z.txt", bytes.NewReader(diff)))
		}
	}
	status := o.getGitStatusText(dir)
	if len(status) > 0 {
		options = append(options,
			xcp.WithFileFromReader("git_status", "git-status-z.txt", bytes.NewReader(status)))
	}
	return options
}

func (o *UploadOpts) getPRDIffText(dir string) []byte {
	buf := &bytes.Buffer{}
	// #nosec G204
	diff := exec.Command("git", "diff", "-z", "--name-status", fmt.Sprintf("%s...HEAD", o.GitPRBaseRef))
	fmt.Fprintf(buf, "# %s\n", strings.Join(diff.Args, " "))
	diff.Dir = dir
	diff.Stdout = buf
	if err := diff.Run(); err != nil {
		log.Warnf("Could not determine PR diffs - {warning:%s}", err)
		return nil
	}
	return buf.Bytes()
}

func (o *UploadOpts) getGitStatusText(dir string) []byte {
	buf := &bytes.Buffer{}
	status := exec.Command("git", "status", "-z")
	fmt.Fprintf(buf, "# %s\n", strings.Join(status.Args, " "))
	status.Dir = dir
	status.Stdout = buf
	if err := status.Run(); err != nil {
		log.Warnf("Could not get git status - {warning:%s}", err)
		return nil
	}
	return buf.Bytes()
}
