package cfnnag

import (
	"fmt"
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
	Templates []string
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "cfn_nag"
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "cfn-nag",
		Short: "Scan cloudformation templates with cfn_nag",
	}
}

func (t *Tool) Register(c *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(c)
	c.Flags().StringSliceVar(&t.Templates, "template", nil,
		"Run cfn_nag on these templates instead of automatically searching for them")
}

func (t *Tool) Run() (*tools.Result, error) {
	files, err := t.findCloudformationFiles()
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no cloudformation templates found")
	}
	d, _ := t.RunDocker(&tools.DockerTool{
		Name:      "cfn_nag",
		Image:     "stelligent/cfn_nag:latest",
		Directory: t.GetDirectory(),
		Args:      append([]string{"--output-format=json"}, files...),
	})
	results, err := jnode.FromJSON(d)
	if err != nil {
		if d != nil {
			os.Stderr.Write(d)
		}
		return nil, err
	}
	violations := jnode.NewArrayNode()
	for _, f := range results.Elements() {
		filename := f.Path("filename").AsText()
		for _, v := range f.Path("file_results").Path("violations").Elements() {
			violations.AppendObject().
				Put("filename", filename).
				Put("id", v.Path("id").AsText()).
				Put("type", v.Path("type").AsText()).
				Put("message", v.Path("message").AsText())
		}
	}
	result := &tools.Result{
		Directory:    t.Directory,
		Data:         results,
		PrintData:    jnode.NewObjectNode().Put("violations", violations),
		PrintPath:    []string{"violations"},
		PrintColumns: []string{"id", "type", "filename", "message"},
	}
	return result, nil
}

func (t *Tool) findCloudformationFiles() ([]string, error) {
	if len(t.Templates) > 0 {
		return t.GetFilesInDirectory(t.Templates)
	}
	return t.GetInventory().CloudformationFiles.Values(), nil
}
