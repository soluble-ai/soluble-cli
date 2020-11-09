package checkov

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

type Tool struct {
	Directory string
}

var _ tools.RunsInDirectory = &Tool{}

func (t *Tool) Name() string {
	return "checkov"
}

func (t *Tool) SetDirectory(dir string) {
	t.Directory = dir
}

func (t *Tool) Run() (*tools.Result, error) {
	if err := tools.HasDocker(); err != nil {
		return nil, err
	}

	// use absolute path for docker volume mapping
	absPath, _ := filepath.Abs(t.Directory)

	// #nosec G204
	c := exec.Command("docker", "run", "-v", fmt.Sprintf("%s:%s", absPath, "/tf"),
		"bridgecrew/checkov", "-d", "/tf", "-o", "json", "-s")
	log.Infof("Running {primary:%s}", strings.Join(c.Args, " "))
	c.Stderr = os.Stderr
	dat, err := c.Output()
	if err != nil {
		if dat != nil {
			_, _ = os.Stderr.Write(dat)
		}
		return nil, err
	}
	n, err := jnode.FromJSON(dat)
	if err != nil {
		return nil, err
	}
	result := &tools.Result{
		Directory: t.Directory,
		Data:      n,
		Values: map[string]string{
			"CHECKOV_VERSION": n.Path("summary").Path("checkov_version").AsText(),
		},
		PrintPath: []string{"soluble_summary"},
		PrintColumns: []string{
			"check_id", "check_result", "file_path", "line", "check_name",
		},
	}
	// checkov has passed_checks and failed_checks so we'll combine them
	// into a summary that we can print out
	s := n.PutArray("soluble_summary")
	processChecks(result, s, n.Path("results").Path("passed_checks"))
	processChecks(result, s, n.Path("results").Path("failed_checks"))
	return result, nil
}

func processChecks(result *tools.Result, s, checks *jnode.Node) {
	for _, n := range checks.Elements() {
		filePath := n.Path("file_path").AsText()
		if len(filePath) > 0 && filePath[0] == '/' {
			filePath = filePath[1:]
		}
		s.AppendObject().Put("check_id", n.Path("check_id").AsText()).
			Put("check_result", n.Path("check_result").Path("result").AsText()).
			Put("file_path", filePath).
			Put("line", n.Path("file_line_range").Get(0).AsInt()).
			Put("check_name", n.Path("check_name").AsText())
	}
}
