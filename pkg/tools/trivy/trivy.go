package trivy

import (
	"io/ioutil"
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
	Image string
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "trivy"
}

func (t *Tool) Run() (*tools.Result, error) {
	m := download.NewManager()
	d, err := m.Install(&download.Spec{
		URL: "github.com/aquasecurity/trivy",
	})
	if err != nil {
		return nil, err
	}
	outfile, err := tempfile()
	if err != nil {
		return nil, err
	}
	program := filepath.Join(d.Dir, "trivy")
	scan := exec.Command(program, "image", "--format", "json", "--output", outfile, t.Image)
	log.Infof("Running {info:%s}", strings.Join(scan.Args, " "))
	scan.Stderr = os.Stderr
	scan.Stdout = os.Stdout
	err = scan.Run()
	if err != nil {
		return nil, err
	}
	dat, err := ioutil.ReadFile(outfile)
	if err != nil {
		return nil, err
	}
	n, err := jnode.FromJSON(dat)
	if err != nil {
		return nil, err
	}
	return &tools.Result{
		Data: n.Get(0),
		Values: map[string]string{
			"TRIVY_VERSION": d.Version,
			"IMAGE":         t.Image,
		},
		PrintPath:    []string{"Vulnerabilities"},
		PrintColumns: []string{"PkgName", "VulnerabilityID", "Severity", "InstalledVersion", "FixedVersion", "Title"},
	}, nil
}

func tempfile() (name string, err error) {
	var f *os.File
	f, err = ioutil.TempFile("", "trivy*")
	if err != nil {
		return
	}
	name = f.Name()
	f.Close()
	return
}
