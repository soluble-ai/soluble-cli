package secrets

import (
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "secrets"
}

func (t *Tool) Run() (*tools.Result, error) {
	d, _ := t.RunDocker(&tools.DockerTool{
		Name:      "soluble-secrets",
		Image:     "gcr.io/soluble-repo/soluble-secrets:latest",
		Directory: t.GetDirectory(),
	})
	results, err := jnode.FromJSON(d)
	if err != nil {
		if d != nil {
			os.Stderr.Write(d)
		}
		return nil, err
	}

	result := t.parseResults(results)

	return result, nil
}

func (t *Tool) parseResults(results *jnode.Node) *tools.Result {
	output := jnode.NewArrayNode()
	findings := []*assessments.Finding{}
	for k, v := range results.Path("results").Entries() {
		if t.IsExcluded(k) {
			continue
		}
		if v.IsArray() {
			for _, p := range v.Elements() {
				p.Put("file_name", k)
				output.Append(p)
				findings = append(findings, &assessments.Finding{
					FilePath: k,
					Line:     p.Path("line_number").AsInt(),
					Title:    p.Path("type").AsText(),
					Tool: map[string]string{
						"is_verified": p.Path("is_verified").AsText(),
					},
				})
			}
		}
	}

	n := jnode.NewObjectNode()
	n.Put("results", output)

	result := &tools.Result{
		Directory:    t.Directory,
		Data:         n,
		Findings:     findings,
		PrintColumns: []string{"filePath", "line", "title", "tool.is_verified"},
	}
	return result
}
