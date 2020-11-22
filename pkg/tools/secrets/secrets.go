package secrets

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

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "secrets"
}

func (t *Tool) Run() (*tools.Result, error) {
	if err := tools.HasDocker(); err != nil {
		return nil, err
	}

	// #nosec G204
	c := exec.Command("docker", "run", "--volume", fmt.Sprintf("%s:%s:ro", t.GetDirectory(), "/repo"),
		"gcr.io/soluble-repo/soluble-secrets:latest")
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
