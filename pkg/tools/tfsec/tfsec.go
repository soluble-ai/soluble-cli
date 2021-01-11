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
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "tfsec"
}

func (t *Tool) Run() (*tools.Result, error) {
	// versions past v0.30.0 seem broken?
	d, err := t.InstallTool(&download.Spec{
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
		Directory: t.GetDirectory(),
		PrintPath: []string{"results"},
		PrintColumns: []string{
			"rule_id",
			"severity",
			"location.filename",
			"location.start_line",
			"description",
		},
	}
	// #nosec G204
	c := exec.Command(d.GetExePath("tfsec-tfsec"), "-f", "json", ".")
	c.Dir = t.GetDirectory()
	c.Stderr = os.Stderr
	log.Infof("Running {primary:%s}", strings.Join(c.Args, " "))
	output, err := c.Output()
	if util.ExitCode(err) == 1 {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	n, err := jnode.FromJSON(output)
	if err != nil {
		return nil, err
	}
	dir := t.GetDirectory()
	if !filepath.IsAbs(dir) {
		// tfsec reports absolute paths which we have to convert to
		// relative paths
		dir, err = filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("could not determine absolute path of %s: %w", dir, err)
		}
	}
	results := n.Path("results")
	if results.Size() > 0 {
		for _, r := range n.Path("results").Elements() {
			loc := r.Path("location")
			filename := loc.Path("filename").AsText()
			if filename != "" && filepath.IsAbs(filename) {
				if f, err := filepath.Rel(dir, filename); err == nil {
					loc.Put("filename", f)
				}
			}
		}
		results = util.RemoveJNodeElementsIf(results, func(e *jnode.Node) bool {
			return t.IsExcluded(e.Path("location").Path("filename").AsText())
		})
		result.Data = jnode.NewObjectNode().Put("results", results)
	}
	return result, nil
}
