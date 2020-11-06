package cfnpythonlint

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

type Tool struct {
	Directory string
}

var _ tools.RunsInDirectory = &Tool{}

func (t *Tool) Name() string {
	return "cfn-python-lint"
}

func (t *Tool) SetDirectory(dir string) {
	t.Directory = dir
}

func (t *Tool) Run() (*tools.Result, error) {
	if err := tools.HasDocker(); err != nil {
		return nil, err
	}
	// #nosec G204
	c := exec.Command("docker", "run", "--volume", fmt.Sprintf("%s:%s:ro", t.Directory, "/data"), "--detach",
		"gcr.io/soluble-repo/soluble-cfn-lint:latest",
		"/data/*.yaml", "/data/*.yml", "/data/*.json", "/data/*.template")
	log.Infof("Running {primary:%s}", strings.Join(c.Args, " "))
	c.Stderr = os.Stderr
	d, err := c.Output()
	spew.Dump(string(d))
	if err != nil {
		if d != nil {
			os.Stderr.Write(d)
		}
	}
	n, err := jnode.FromJSON(d)
	if err != nil {
		return nil, err
	}
	result := &tools.Result{
		Directory: t.Directory,
		Data:      n,
	}
	return result, nil
}
