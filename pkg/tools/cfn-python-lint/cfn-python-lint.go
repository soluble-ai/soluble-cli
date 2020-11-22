package cfnpythonlint

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
}

func (t *Tool) Name() string {
	return "cfn-python-lint"
}

func (t *Tool) Run() (*tools.Result, error) {
	if err := tools.HasDocker(); err != nil {
		return nil, err
	}
	// #nosec G204
	c := exec.Command("docker", "run", "--volume", fmt.Sprintf("%s:%s:ro", t.GetDirectory(), "/data"),
		"gcr.io/soluble-repo/soluble-cfn-lint:latest",
		"/data/**/*.yaml", "/data/**/*.yml", "/data/**/*.json", "/data/**/*.template")
	log.Infof("Running {primary:%s}", strings.Join(c.Args, " "))
	c.Stderr = os.Stderr
	d, _ := c.Output()
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
