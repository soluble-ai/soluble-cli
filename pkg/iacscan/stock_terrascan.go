package iacscan

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"gopkg.in/yaml.v2"
)

type StockTerrascan struct {
	Directory string
}

func (t *StockTerrascan) Run() (*jnode.Node, error) {
	m := download.NewManager()
	d, err := m.InstallGithubRelease("accurics", "terrascan", "")
	if err != nil {
		return nil, err
	}
	program := filepath.Join(d.Dir, "terrascan")
	init := exec.Command(program, "init")
	init.Stdout = os.Stderr
	init.Stderr = os.Stderr
	log.Infof("%s", init.Args)
	if err = init.Run(); err != nil {
		return nil, fmt.Errorf("terrascan init failed: %w", err)
	}
	scan := exec.Command(program, "scan", "-t", "aws", "-d", t.Directory)
	scan.Stderr = os.Stderr
	out := &bytes.Buffer{}
	scan.Stdout = out
	scan.Stderr = os.Stderr
	log.Infof("%s", scan.Args)
	_ = scan.Run()
	var result map[string]interface{}
	if err = yaml.Unmarshal(out.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("could not parse terrascan output: %w", err)
	}
	return jnode.FromMap(result), nil
}
