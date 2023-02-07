package tools

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-resty/resty/v2"
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
		forkPoint := o.getGitForkPoint(dir)
		if forkPoint != "" {
			options = append(options, api.OptionFunc(func(r *resty.Request) {
				r.SetMultipartField("SOLUBLE_METADATA_GIT_FORK_POINT", "", "", strings.NewReader(forkPoint))
			}))
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
	gsbuf := &bytes.Buffer{}
	gs := exec.Command("git", "status")
	gs.Stdout = gsbuf
	err := gs.Run()
	if err != nil {
		return nil
	}
	log.Warnf("git status foo %s", string(gsbuf.Bytes()))
	// #nosec G204
	diff := exec.Command("git", "diff", "-z", "--name-status", fmt.Sprintf("%s...HEAD", o.GitPRBaseRef))
	fmt.Fprintf(buf, "# %s\n", strings.Join(diff.Args, " "))
	diff.Dir = dir
	diff.Stdout = buf
	if err := diff.Run(); err != nil {
		log.Warnf("Could not determine PR diffs - {warning:%s}", err)
		log.Warnf("Could not determine PR diffs - {warning:%s}", dir)
		log.Warnf("Could not determine PR diffs - {warning:%s}", string(buf.Bytes()))
		return nil
	}
	return buf.Bytes()
}

func (o *UploadOpts) getGitForkPoint(dir string) string {
	buf := &strings.Builder{}
	// #nosec G204
	diff := exec.Command("git", "merge-base", "--fork-point", o.GitPRBaseRef)
	diff.Dir = dir
	diff.Stdout = buf
	if err := diff.Run(); err != nil {
		log.Warnf("Could not determine git fork point - {warning:%s}", err)
		return ""
	}
	return strings.TrimSpace(buf.String())
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
