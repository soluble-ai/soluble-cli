package cfnpythonlint

import (
	"fmt"
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
}

func (t *Tool) Name() string {
	return "cfn-python-lint"
}

func (t *Tool) Run() (*tools.Result, error) {
	d, _ := t.RunDocker(&tools.DockerTool{
		Name:  "cfn-python-lint",
		Image: "gcr.io/soluble-repo/soluble-cfn-lint:latest",
		DockerArgs: []string{
			"--volume", fmt.Sprintf("%s:%s:ro", t.GetDirectory(), "/data"),
		},
		Args: []string{
			"/data/**/*.yaml", "/data/**/*.yml", "/data/**/*.json", "/data/**/*.template",
		},
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
