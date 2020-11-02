package tfsec

import (
	"fmt"
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

var _ tools.RunsInDirectory = &Tool{}

func (t *Tool) Name() string {
	return "tfsec"
}

func (t *Tool) SetDirectory(dir string) {
	t.Directory = dir
}

func (t *Tool) Run() (*tools.Result, error) {
	m := download.NewManager()
	// versions past v0.30.0 seem broken?
	d, err := m.Install(&download.Spec{
		URL:              "github.com/tfsec/tfsec",
		RequestedVersion: "v0.30.0",
	})
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
	dir := t.Directory
	if !filepath.IsAbs(dir) {
		// tfsec reports absolute paths which we have to convert to
		// relative paths
		dir, err = filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("could not determine absolute path of %s: %w", dir, err)
		}
	}
	for _, r := range n.Path("results").Elements() {
		loc := r.Path("location")
		r.Put("line", loc.Path("start_line").AsInt())
		file, _ := filepath.Rel(dir, loc.Path("filename").AsText())
		r.Put("file", file)
	}
	result.Data = n
	return result, nil
}
