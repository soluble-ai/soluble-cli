package secrets

import (
	"fmt"
	"os"

	"github.com/soluble-ai/go-jnode"
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
		Name:  "soluble-secrets",
		Image: "gcr.io/soluble-repo/soluble-secrets:latest",
		DockerArgs: []string{
			"--volume", fmt.Sprintf("%s:%s:ro", t.GetDirectory(), "/repo"),
		},
	})
	results, err := jnode.FromJSON(d)
	if err != nil {
		if d != nil {
			os.Stderr.Write(d)
		}
		return nil, err
	}

	output := jnode.NewArrayNode()
	for k, v := range results.Path("results").Entries() {
		if v.IsArray() {
			for _, p := range v.Elements() {
				p.Put("file_name", k)
				output.Append(p)
			}
		}
	}

	n := jnode.NewObjectNode()
	n.Put("results", output)

	result := &tools.Result{
		Directory:    t.Directory,
		Data:         n,
		PrintPath:    []string{"results"},
		PrintColumns: []string{"file_name", "type", "line_number", "is_verified"},
	}
	return result, nil
}
