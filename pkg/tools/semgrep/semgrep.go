package semgrep

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
	RulesDir string
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "semgrep"
}

func (t *Tool) SetDirectory(dir string) {
	t.Directory = dir
}

func (t *Tool) GetRulesDir() (string, error) {
	if t.RulesDir != "" {
		return t.RulesDir, nil
	}
	dir, err := ioutil.TempDir("", "soluble-semgrep-tmp*")
	if err != nil {
		log.Errorf("unable to create tmpdir: %w", err)
	}
	t.RulesDir = dir
	// TODO: pull concatenated rules in from API.
	// Until then, using a hack below
	var ruleConfig string = `rules:
- id: k8s-repos
  patterns:
  - pattern: |
      image:...
  - pattern-not: |
      image: gcr.io/some-whitelisted-repo...
  languages: [generic]
  message: |
    $IMAGENAME,$PATH
  severity: WARNING
`
	// Remove code above once rules are in place.
	err = ioutil.WriteFile(filepath.Join(t.RulesDir, "semgrep-combined.yml"), []byte(ruleConfig), 0o777)
	if err != nil {
		return "", err
	}
	return t.RulesDir, nil
}

func (t *Tool) Run() (*tools.Result, error) {
	rulesDir, err := t.GetRulesDir()
	if err != nil {
		return nil, fmt.Errorf("unable to get rules dir: %w", err)
	}
	dat, err := t.RunDocker(&tools.DockerTool{
		Name:  "semgrep",
		Image: "gcr.io/soluble-repo/semgrep:latest",
		DockerArgs: []string{
			"-v", fmt.Sprintf("%s:%s:ro", t.GetDirectory(), "/data"),
			"-v", fmt.Sprintf("%s:%s:ro", rulesDir, "/rules"),
		},
		Args: []string{
			"--config=/rules/semgrep-combined.yml", "--json",
			//"--output=/tmp/results.json",
			"/data/",
		},
	})
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

	output := jnode.NewObjectNode()
	data := output.PutArray("data")

	if n.IsArray() {
		for _, e := range n.Elements() {
			data = data.Append(e)
		}
	} else {
		data.Append(n)
	}

	result := &tools.Result{
		Directory: t.Directory,
		Data:      output,
		// Values:    map[string]string{},
		// PrintPath: []string{},
		// PrintColumns: []string{},
	}

	return result, nil
}
