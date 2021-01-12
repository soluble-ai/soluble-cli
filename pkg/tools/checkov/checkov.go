package checkov

import (
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts

	extraArgs []string
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "checkov"
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "checkov",
		Short: "Run checkov",
		Example: `# Any additional args after -- are passed through to checkov, eg:
... checkov -- --help`,
		Args: func(cmd *cobra.Command, args []string) error {
			t.extraArgs = args
			return nil
		},
	}
}

func (t *Tool) Run() (*tools.Result, error) {
	dat, err := t.RunDocker(&tools.DockerTool{
		Name:      "checkov",
		Image:     "gcr.io/soluble-repo/checkov:latest",
		Directory: t.GetDirectory(),
		Args: append([]string{
			"-d", ".", "-o", "json", "-s",
		}, t.extraArgs...),
	})
	if err != nil {
		if dat != nil {
			_, _ = os.Stderr.Write(dat)
		}
		return nil, err
	}
	n, err := jnode.FromJSON(dat)
	if err != nil {
		_, _ = os.Stderr.Write(dat)
		return nil, err
	}

	// checkov runs various types of check such as kubernetes, terraform etc if the folder has
	// different types of them in the same folder the result will be an array with "check_type"
	// of each element indicating if it was kubernetes, terraform, etc
	data := jnode.NewArrayNode()
	var checkovVersion string
	if n.IsArray() {
		for _, e := range n.Elements() {
			data = data.Append(e)
		}
		checkovVersion = n.Get(0).Path("summary").Path("checkov_version").AsText()
	} else {
		checkovVersion = n.Path("summary").Path("checkov_version").AsText()
		data.Append(n)
	}

	result := &tools.Result{
		Directory: t.Directory,
		Values: map[string]string{
			"CHECKOV_VERSION": checkovVersion,
		},
		Data:      jnode.NewObjectNode().Put("data", data),
		PrintData: jnode.NewArrayNode(),
		PrintPath: []string{},
		PrintColumns: []string{
			"check_id", "check_result", "check_type", "file_path", "line", "check_name",
		},
	}
	t.processResults(result, data)

	return result, nil
}

func (t *Tool) processResults(result *tools.Result, data *jnode.Node) {
	for _, e := range data.Elements() {
		checkType := e.Path("check_type").AsText()
		results := e.Path("results")
		passedChecks := t.processChecks(result, results.Path("passed_checks"), checkType)
		failedChecks := t.processChecks(result, results.Path("failed_checks"), checkType)
		updateChecks(results, "passed_checks", passedChecks)
		updateChecks(results, "failed_checks", failedChecks)
	}
}

func updateChecks(results *jnode.Node, name string, checks *jnode.Node) {
	if checks.Size() == 0 {
		results.Remove(name)
	} else {
		results.Put(name, checks)
	}
}

func (t *Tool) processChecks(result *tools.Result, checks *jnode.Node, checkType string) *jnode.Node {
	for _, n := range checks.Elements() {
		filePath := n.Path("file_path").AsText()
		if len(filePath) > 0 && filePath[0] == '/' {
			filePath = filePath[1:]
			n.Put("file_path", filePath)
		}
	}
	checks = util.RemoveJNodeElementsIf(checks, func(e *jnode.Node) bool {
		return t.IsExcluded(e.Path("file_path").AsText())
	})
	for _, n := range checks.Elements() {
		result.PrintData.AppendObject().Put("check_id", n.Path("check_id").AsText()).
			Put("check_result", n.Path("check_result").Path("result").AsText()).
			Put("file_path", n.Path("file_path")).
			Put("line", n.Path("file_line_range").Get(0).AsInt()).
			Put("check_name", n.Path("check_name").AsText()).
			Put("check_type", checkType)
	}
	return checks
}
