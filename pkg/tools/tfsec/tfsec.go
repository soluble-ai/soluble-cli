package tfsec

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

type Tool struct {
	Directory string
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "tfsec"
}

func (t *Tool) Run() (*tools.Result, error) {
	m := download.NewManager()
	// versions past v0.30.0 seem broken?
	d, err := m.InstallGithubRelease("tfsec", "tfsec", "v0.30.0")
	if err != nil {
		return nil, err
	}
	result := &tools.Result{
		Values: map[string]string{
			"TFSEC_VERSION": d.Version,
		},
		Directory: t.Directory,
		PrintPath: []string{"results"},
		PrintColumns: []string{
			"rule_id",
			"severity",
			"file",
			"line",
			"description",
		},
	}
	// #nosec G204
	c := exec.Command(filepath.Join(d.Dir, "tfsec-tfsec"), "-f", "json", ".")
	c.Dir = t.Directory
	c.Stderr = os.Stderr
	log.Infof("Running {primary:%s}", strings.Join(c.Args, " "))
	output, err := c.Output()
	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	n, err := jnode.FromJSON(output)
	if err != nil {
		return nil, err
	}
	for _, r := range n.Path("results").Elements() {
		r.Put("line", r.Path("location").Path("start_line").AsInt())
		file, _ := filepath.Rel(t.Directory, r.Path("location").Path("filename").AsText())
		r.Put("file", file)
		result.AddFile(file)
	}
	result.Data = n
	return result, nil
}
