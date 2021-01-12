package trivy

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.ToolOpts
	Image      string
	ClearCache bool
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "trivy"
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.ToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVarP(&t.Image, "image", "i", "", "The image to scan")
	flags.BoolVarP(&t.ClearCache, "clear-cache", "c", false, "clear image caches and then start scanning")
	_ = cmd.MarkFlagRequired("image")
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "image-scan",
		Short: "Scan a container image",
		Args:  cobra.ArbitraryArgs,
	}
}

func (t *Tool) Run() (*tools.Result, error) {
	d, err := t.InstallTool(&download.Spec{
		URL: "github.com/aquasecurity/trivy",
	})
	if err != nil {
		return nil, err
	}
	outfile, err := tools.TempFile("trivy*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(outfile)
	program := d.GetExePath("trivy")
	if t.ClearCache {
		err := runCommand(program, "image", "--clear-cache")
		if err != nil {
			return nil, err
		}
	}

	// Generate params for the scanner
	args := []string{"image", "--format", "json", "--output", outfile}
	// specify the image to scan at the end of params
	args = append(args, t.Image)

	err = runCommand(program, args...)
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

func runCommand(program string, args ...string) error {
	scan := exec.Command(program, args...)
	log.Infof("Running {info:%s}", strings.Join(scan.Args, " "))
	scan.Stderr = os.Stderr
	scan.Stdout = os.Stdout
	err := scan.Run()
	if err != nil {
		return err
	}
	return nil
}
