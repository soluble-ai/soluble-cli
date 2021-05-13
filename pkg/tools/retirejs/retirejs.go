package retirejs

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Interface = (*Tool)(nil)

func (t *Tool) Name() string { return "retirejs" }

func (t *Tool) Run() (*tools.Result, error) {
	args := []string{
		"retire", "--exitwith", "0", "--outputformat", "json", "--path", ".",
	}
	var output bytes.Buffer
	_, err := t.RunDocker(&tools.DockerTool{
		Name:                "retirejs",
		Image:               "gcr.io/soluble-repo/soluble-retirejs:latest",
		DefaultNoDockerName: "retire",
		Directory:           t.GetDirectory(),
		Args:                args,
		Stderr:              &output, // retirejs writes to stderr, sigh
		Stdout:              os.Stderr,
	})
	if err != nil {
		_, _ = os.Stderr.Write(output.Bytes())
		return nil, err
	}
	n, err := jnode.FromJSON(output.Bytes())
	if err != nil {
		_, _ = os.Stderr.Write(output.Bytes())
		return nil, err
	}
	result := t.parseResults(n)
	return result, nil
}

func (t *Tool) parseResults(results *jnode.Node) *tools.Result {
	findings := assessments.Findings{}
	for _, data := range results.Path("data").Elements() {
		file := data.Path("file").AsText()
		if filepath.IsAbs(file) {
			runDir := t.GetDockerRunDirectory()
			if !strings.HasPrefix(file, runDir) {
				panic(fmt.Sprintf("%s does not start with %s", file, runDir))
			}
			file = file[len(runDir)+1:]
		}
		if t.IsExcluded(file) {
			continue
		}
		for _, r := range data.Path("results").Elements() {
			for _, v := range r.Path("vulnerabilities").Elements() {
				findings = append(findings, &assessments.Finding{
					FilePath: file,
					Tool: map[string]string{
						"component":  r.Path("component").AsText(),
						"version":    r.Path("version").AsText(),
						"severity":   v.Path("severity").AsText(),
						"identifier": getVulnIdentifier(v),
					},
				})
			}
		}
	}
	dataArray := results.Path("data")
	dataArray = util.RemoveJNodeElementsIf(dataArray, func(n *jnode.Node) bool {
		return t.IsExcluded(n.Path("file").AsText())
	})
	results.Put("data", dataArray)
	result := &tools.Result{
		Data:         results,
		Findings:     findings,
		PrintColumns: []string{"tool.component", "tool.version", "tool.severity", "tool.identifier", "filePath"},
	}
	return result
}

func getVulnIdentifier(v *jnode.Node) string {
	for id, valn := range v.Path("identifiers").Entries() {
		var val string
		if valn.IsArray() {
			if valn.Size() > 0 {
				val = valn.Get(0).AsText()
			}
		} else {
			val = valn.AsText()
		}
		if strings.HasPrefix(val, id) {
			return val
		}
		return fmt.Sprintf("%s-%s", id, val)
	}
	return ""
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "retirejs",
		Short: "Run retirejs to find node/javascript dependencies that should be retired",
	}
}
