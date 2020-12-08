package checkov

import (
	"os"
	"os/exec"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "checkov"
}

func (t *Tool) SetDirectory(dir string) {
	t.Directory = dir
}

func (t *Tool) Run() (*tools.Result, error) {
	log.Infof("Installing checkov")
	// #nosec G204
	c := exec.Command("pip3", "install", "checkov")
	c.Stderr = os.Stderr
	_, err := c.Output()
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok || ee.ExitCode() != 0 {
			return nil, err
		}
	}
	// #nosec G204
	c = exec.Command("checkov", "-o", "json", "-s", "-d", t.GetDirectory())
	c.Stderr = os.Stderr
	o, err := c.Output()
	if err != nil {
		return nil, err
	}
	n, err := jnode.FromJSON(o)
	if err != nil {
		return nil, err
	}
	// checkov runs various types of check such as kubernetes, terraform etc if the folder has
	// different types of them in the same folder the result will be an array
	output := jnode.NewObjectNode()
	data := output.PutArray("data")

	// checkov has passed_checks and failed_checks so we'll combine them
	s := output.PutArray("soluble_summary")
	var checkovVersion string
	if n.IsArray() {
		for _, e := range n.Elements() {
			data = data.Append(e)
		}
		checkovVersion = n.Get(0).Path("summary").Path("checkov_version").AsText()
	} else {
		checkovVersion = n.Path("summary").Path("checkov_version").AsText()
		data.Append(n)
	}

	result := &tools.Result{
		Directory: t.Directory,
		Data:      output,
		Values: map[string]string{
			"CHECKOV_VERSION": checkovVersion,
		},
		PrintPath: []string{"soluble_summary"},
		PrintColumns: []string{
			"check_id", "check_result", "check_type", "file_path", "line", "check_name",
		},
	}

	for _, e := range output.Path("data").Elements() {
		checkType := e.Path("check_type").AsText()
		processChecks(result, s, e.Path("results").Path("passed_checks"), checkType)
		processChecks(result, s, e.Path("results").Path("failed_checks"), checkType)
	}

	return result, nil
}

func processChecks(result *tools.Result, s, checks *jnode.Node, checkType string) {
	for _, n := range checks.Elements() {
		filePath := n.Path("file_path").AsText()
		if len(filePath) > 0 && filePath[0] == '/' {
			filePath = filePath[1:]
		}
		s.AppendObject().Put("check_id", n.Path("check_id").AsText()).
			Put("check_result", n.Path("check_result").Path("result").AsText()).
			Put("file_path", filePath).
			Put("line", n.Path("file_line_range").Get(0).AsInt()).
			Put("check_name", n.Path("check_name").AsText()).
			Put("check_type", checkType)
	}
}
