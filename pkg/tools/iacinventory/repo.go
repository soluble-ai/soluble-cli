package iacinventory

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/repotree"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
	"github.com/spf13/cobra"
)

type Repo struct {
	tools.ToolOpts
	tools.DirectoryOpt
	tools.UploadOpt
	Details bool
}

var _ tools.Simple = (*Repo)(nil)

func (*Repo) Name() string {
	return "repo-inventory"
}

func (*Repo) IsNonAssessment() bool {
	return true
}

func (r *Repo) Register(cmd *cobra.Command) {
	r.ToolOpts.Register(cmd)
	r.DirectoryOpt.Register(cmd)
	r.UploadOpt.Register(cmd)
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
	if !r.Details {
		tree.Files = nil
	}
	n, err := print.ToResult(tree)
	if err != nil {
		return err
	}
	r.PrintResult(n)
	if r.UploadEnabled {
		values := r.GetStandardXCPValues()
		options := []api.Option{
			xcp.WithFileFromReader("tree", "tree.json", bytes.NewReader(treeDat)),
		}
		_, err := r.GetAPIClient().XCPPost(r.GetOrganization(), "repo-tree", nil, values, options...)
		if err != nil {
			return err
		}
	}
	return nil
}
