package tfsec

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
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

	result := t.parseResults(n)
	result.AddValue("TFSEC_VERSION", d.Version)
	return result, nil
}

func (t *Tool) parseResults(n *jnode.Node) *tools.Result {
	dir := t.GetDirectory()
	results := n.Path("results")
	var findings []*assessments.Finding
	if results.Size() > 0 {
		for _, r := range n.Path("results").Elements() {
			loc := r.Path("location")
			filename := loc.Path("filename").AsText()
			if filename != "" && filepath.IsAbs(filename) {
				f, err := filepath.Rel(dir, filename)
				if err == nil {
					loc.Put("filename", f)
					filename = f
				}
			}
			findings = append(findings, &assessments.Finding{
				FilePath:    filename,
				Line:        r.Path("location").Path("start_line").AsInt(),
				Description: r.Path("description").AsText(),
				Tool: map[string]string{
					"severity": r.Path("severity").AsText(),
					"rule_id":  r.Path("rule_id").AsText(),
				},
			})
		}
		results = util.RemoveJNodeElementsIf(results, func(e *jnode.Node) bool {
			return t.IsExcluded(e.Path("location").Path("filename").AsText())
		})
		n.Put("results", results)
	}
	return &tools.Result{
		Directory: t.GetDirectory(),
		Data:      n,
		Findings:  findings,
		PrintColumns: []string{
			"tool.rule_id",
			"tool.severity",
			"filePath",
			"line",
			"description",
		},
	}
}
