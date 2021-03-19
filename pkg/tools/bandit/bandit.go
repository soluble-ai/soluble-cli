package bandit

import (
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Interface = (*Tool)(nil)

func (t *Tool) Name() string {
	return "bandit"
}

func (t *Tool) Run() (*tools.Result, error) {
	args := []string{
		"--exit-zero", "-f", "json", "-r", ".",
	}
	dat, err := t.RunDocker(&tools.DockerTool{
		Name:      "bandit",
		Image:     "gcr.io/soluble-repo/soluble-bandit:latest",
		Directory: t.GetDirectory(),
		Args:      args,
	})
	if err != nil {
		if dat != nil {
			_, _ = os.Stderr.Write(dat)
		}
		return nil, err
	}
	n, err := jnode.FromJSON(dat)
	if err != nil {
		_, _ = os.Stderr.Write(dat)
		return nil, err
	}
	result := parseResults(n)
	return result, nil
}

func parseResults(results *jnode.Node) *tools.Result {
	findings := assessments.Findings{}
	for _, r := range results.Path("results").Elements() {
		findings = append(findings, &assessments.Finding{
			FilePath: r.Path("filename").AsText(),
			Line:     r.Path("line_number").AsInt(),
			Tool: map[string]string{
				"severity":   r.Path("issue_severity").AsText(),
				"confidence": r.Path("issue_confidence").AsText(),
				"id":         r.Path("test_id").AsText(),
				"name":       r.Path("test_name").AsText(),
			},
		})
	}
	result := &tools.Result{
		Data:         results,
		Findings:     findings,
		PrintColumns: []string{"tool.id", "tool.name", "tool.severity", "tool.confidence", "filePath", "line"},
	}
	return result
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "bandit",
		Short: "Run bandit to identify security flaws in python",
	}
}
