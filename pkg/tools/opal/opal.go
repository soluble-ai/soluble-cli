package opal

import (
	"os"
	"os/exec"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Single = (*Tool)(nil)

func (t *Tool) Name() string {
	return "opal"
}

func (t *Tool) Run() (*tools.Result, error) {
	result := &tools.Result{
		Directory:   t.GetDirectory(),
		IACPlatform: "terraform",
	}
	d, err := t.InstallTool(&download.Spec{Name: "opal"})
	if err != nil {
		return nil, err
	}
	customPoliciesDir, err := t.GetCustomPoliciesDir()
	if err != nil {
		return nil, err
	}
	args := []string{"run", "--format", "json"}
	if customPoliciesDir != "" {
		args = append(args, "--include", customPoliciesDir)
	}
	args = append(args, ".")
	// #nosec G204
	c := exec.Command(d.GetExePath("opal"), args...)
	c.Dir = t.GetDirectory()
	c.Stderr = os.Stderr
	exec := t.ExecuteCommand(c)
	result.ExecuteResult = exec
	if !exec.ExpectExitCode(0, 1) {
		return result, nil
	}
	n, ok := exec.ParseJSON()
	if !ok {
		return result, nil
	}
	t.parseResults(result, n)
	return result, nil
}

func (t *Tool) parseResults(result *tools.Result, n *jnode.Node) {
	for _, rr := range n.Path("rule_results").Elements() {
		loc := rr.Path("source_location").Get(0)
		result.Findings = append(result.Findings, &assessments.Finding{
			Severity: rr.Path("rule_severity").AsText(),
			Pass:     rr.Path("rule_result").AsText() == "PASS",
			FilePath: loc.Path("path").AsText(),
			Line:     loc.Path("line").AsInt(),
			Title:    rr.Path("rule_summary").AsText(),
		})
	}
}
