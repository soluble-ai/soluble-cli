package secrets

import (
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
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

func (t *Tool) excludeResults(results *jnode.Node) {
	rs := results.Path("results")
	if rs.Size() > 0 {
		results.Put("results", util.RemoveJNodeEntriesIf(rs, func(k string, v *jnode.Node) bool {
			return t.IsExcluded(k)
		}))
	}
}

func (t *Tool) parseResults(results *jnode.Node) *tools.Result {
	t.excludeResults(results)
	findings := []*assessments.Finding{}
	for k, v := range results.Path("results").Entries() {
		if v.IsArray() {
			for _, p := range v.Elements() {
				p.Put("file_name", k)
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
	result := &tools.Result{
		Directory:    t.Directory,
		Data:         results,
		Findings:     findings,
		PrintColumns: []string{"filePath", "line", "title", "tool.is_verified"},
	}
	return result
}
