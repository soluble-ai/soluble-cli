package iacinventory

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"

	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/repotree"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
	"github.com/spf13/cobra"
)

type Repo struct {
	tools.ToolOpts
	tools.DirectoryOpt
	tools.UploadOpts
	Details bool
}

var _ tools.Simple = (*Repo)(nil)

func (*Repo) Name() string {
	return "repo-inventory"
}

func (*Repo) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "repo-inventory",
		Short: "Inventory a git repository and extract infrastructure-as-code metadata",
	}
}

func (r *Repo) Register(cmd *cobra.Command) {
	r.ToolOpts.Register(cmd)
	r.DirectoryOpt.Register(cmd)
	r.UploadOpts.Register(cmd)
	flags := cmd.Flags()
	flags.BoolVar(&r.Details, "details", false, "Print out the file tree along with a summary")
}

func (r *Repo) Validate() error {
	if err := r.DirectoryOpt.Validate(&r.ToolOpts); err != nil {
		return err
	}
	if err := r.ToolOpts.Validate(); err != nil {
		return err
	}
	if r.RepoRoot == "" {
		return fmt.Errorf("this command must be run within a git repository")
	}
	return nil
}

func (r *Repo) Run() error {
	tree, err := repotree.Do(r.GetDirectory())
	if err != nil {
		return err
	}
	var treeDat []byte
	if r.UploadEnabled {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", " ")
		if err := enc.Encode(tree); err != nil {
			return err
		}
		treeDat = buf.Bytes()
	}
	if r.UploadEnabled {
		values := r.GetStandardXCPValues()
		gzdat := &bytes.Buffer{}
		gz := gzip.NewWriter(gzdat)
		if _, err := io.Copy(gz, bytes.NewReader(treeDat)); err != nil {
			return err
		}
		if err := gz.Flush(); err != nil {
			return err
		}
		options := []api.Option{
			xcp.WithCIEnv(r.GetDirectory()),
			xcp.WithFileFromReader("tree", "tree.json.gz", gzdat),
		}
		options = r.AppendUploadOptions(r.GetDirectory(), options)
		log.Infof("Uploading {info:%s} of compressed tree data", util.Size(uint64(gzdat.Len())))
		api, err := r.GetAPIClient()
		if err != nil {
			return err
		}
		_, err = api.XCPPost("repo-tree", nil, values, options...)
		if err != nil {
			return err
		}
	}
	if !r.Details {
		tree.Files = nil
	}
	n, err := print.ToResult(tree)
	if err != nil {
		return err
	}
	r.PrintResult(n)
	return nil
}
