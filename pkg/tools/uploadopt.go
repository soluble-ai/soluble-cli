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
	UploadEnabled bool
	GitPRBaseRef  string
}

func (o *UploadOpt) Register(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.BoolVar(&o.UploadEnabled, "upload", false, "Upload results")
	flags.StringVar(&o.GitPRBaseRef, "git-pr-base-ref", "", "The pull request base `ref` for diffs.")
}

func (o *UploadOpt) AddPRDiffsUpload(result *Result) {
	if result.Directory == "" {
		return
	}
	opt := o.GetPRDiffUploadOption(result.Directory)
	if opt != nil {
		result.AddUploadOption(opt)
	}
}

func (o *UploadOpt) GetPRDiffUploadOption(dir string) api.Option {
	diff := o.getPRDIffText(dir)
	if diff != nil {
		return xcp.WithFileFromReader("git_pr_diffs", "git-pr-diffs.txt", bytes.NewReader(diff))
	}
	return nil
}

func (o *UploadOpt) getPRDIffText(dir string) []byte {
	if o.GitPRBaseRef == "" {
		return nil
	}
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
