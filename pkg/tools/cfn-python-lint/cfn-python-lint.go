package cfnpythonlint

import (
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
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

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "cfn-python-lint",
		Short: "Scan cloudformation templates with cfn-python-lint",
	}
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
	result := parseResults(results)
	result.Directory = t.GetDirectory()
	return result, nil
}

func parseResults(results *jnode.Node) *tools.Result {
	findings := assessments.Findings{}
	for _, r := range results.Elements() {
		findings = append(findings, &assessments.Finding{
			FilePath: r.Path("Filename").AsText(),
			Line:     r.Path("Location").Path("Start").Path("LineNumber").AsInt(),
			Tool: map[string]string{
				"Level":   r.Path("Level").AsText(),
				"Message": util.TruncateRight(r.Path("Message").AsText(), 100),
				"Rule_Id": r.Path("Rule").Path("Id").AsText(),
			},
		})
	}
	result := &tools.Result{
		Data:         results,
		Findings:     findings,
		PrintColumns: []string{"tool.Rule_Id", "tool.Level", "filePath", "line", "tool.Message"},
	}
	return result
}

func (t *Tool) findCloudformationFiles() ([]string, error) {
	if len(t.Templates) > 0 {
		return t.GetFilesInDirectory(t.Templates)
	}
	return t.GetInventory().CloudformationFiles.Values(), nil
}
