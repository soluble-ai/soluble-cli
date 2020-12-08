package cfnpythonlint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
	Templates []string
}

func (t *Tool) Name() string {
	return "cfn-python-lint"
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(cmd)
	cmd.Flags().StringSliceVar(&t.Templates, "template", nil, "Explicitly specific templates in the form `t1,t2,...`.  May be repeated.  Templates must be relative to --directory.")
}

func (t *Tool) Run() (*tools.Result, error) {
	files, err := t.findCloudformationFiles()
	if err != nil {
		return nil, err
	}
	d, _ := t.RunDocker(&tools.DockerTool{
		Name:      "cfn-python-lint",
		Image:     "gcr.io/soluble-repo/soluble-cfn-lint:latest",
		Directory: t.GetDirectory(),
		Args:      append([]string{"-f", "json"}, files...),
	})
	results, err := jnode.FromJSON(d)
	if err != nil {
		if d != nil {
			os.Stderr.Write(d)
		}
		return nil, err
	}
	n := jnode.NewObjectNode()
	n.Put("results", results)
	result := &tools.Result{
		Directory:    t.Directory,
		Data:         n,
		PrintPath:    []string{"results"},
		PrintColumns: []string{"Rule.Id", "Level", "Filename", "Message"},
	}
	return result, nil
}

func (t *Tool) findCloudformationFiles() ([]string, error) {
	files := []string{}
	if len(t.Templates) > 0 {
		for _, f := range t.Templates {
			rf := f
			if filepath.IsAbs(f) {
				if !strings.HasPrefix(f, t.GetDirectory()) {
					return nil, fmt.Errorf("template file %s must be relative to --directory", f)
				}
			}
			files = append(files, rf)
		}
	} else {
		m := inventory.Do(t.GetDirectory())
		files = m.CloudformationFiles.Values()
	}
	return files, nil
}
