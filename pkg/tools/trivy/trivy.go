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
	Image         string
	IgnoreUnfixed bool
	ClearCache    bool
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
	flags.BoolVarP(&t.IgnoreUnfixed, "ignore-unfixed", "u", false, "display only fixed vulnerabilities")
	_ = cmd.MarkFlagRequired("image")
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "image-scan",
		Short: "Scan a container image",
	}
}

func (t *Tool) Run() (*tools.Result, error) {
	d, err := t.InstallTool(&download.Spec{
		URL: "github.com/aquasecurity/trivy",
	})
	if err != nil {
		return nil, err
	}
	outfile, err := tempfile()
	if err != nil {
		return nil, err
	}
	program := d.GetExePath("trivy")
	if t.ClearCache {
		err := runCommand(program, "image", "--clear-cache")
		if err != nil {
			return nil, err
		}
	}

	// Generate params for the scanner
	args := []string{"image", "--format", "json", "--output", outfile}
	if t.IgnoreUnfixed {
		args = append(args, "--ignore-unfixed")
	}
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

// DirTool implements dependency scanning for trivy.
// However the output from the filesystem scanner is distinctly
// different from the image scanning, and is also invoked from
// a different command.
type DirTool struct {
	Tool
	Dir string
}

func (t *DirTool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "trivy",
		Short: "Scan a local directory",
	}
}

func (t *DirTool) Register(cmd *cobra.Command) {
	t.ToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.BoolVarP(&t.IgnoreUnfixed, "ignore-unfixed", "u", false, "display only fixed vulnerabilities")
	flags.StringVarP(&t.Dir, "directory", "d", ".", "The directory to scan")
}

func (t *DirTool) Run() (*tools.Result, error) {
	d, err := t.InstallTool(&download.Spec{
		URL: "github.com/aquasecurity/trivy",
	})
	if err != nil {
		return nil, err
	}
	outfile, err := tempfile()
	if err != nil {
		return nil, err
	}
	program := d.GetExePath("trivy")
	args := []string{"fs", "--format", "json", "--output", outfile}
	if t.IgnoreUnfixed {
		args = append(args, "--ignore-unfixed")
	}
	args = append(args, ".") // scan the current directory
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

	// If trivy turned up nothing...
	if n.Size() == 0 {
		log.Infof("No vulnerabilities identified by trivy scan")
		return nil, nil
	}

	/*
		// we must merge elements in the array from...

			 [
			  {
			    "Target": ".github/action/package-lock.json",
			    "Type": "npm",
			    "Vulnerabilities": [
			      {
			        "VulnerabilityID": "CVE-2020-15228",
			        "PkgName": "@actions/core",
			        "InstalledVersion": "1.2.4",
				...omitted for brevity...
			        "LastModifiedDate": "2020-12-23T18:32:00Z"
			      }
			    ]
			  },
			  {
			    "Target": "package-lock.json",
			    "Type": "npm",
			    "Vulnerabilities": [
			      {
			        "VulnerabilityID": "CVE-2020-8244",
			        "PkgName": "bl",
			        "InstalledVersion": "4.0.2",
				...omitted for brevity...
			      }
			    ]
			  },
			  ...omitted for brevity...
			 ]


			 to:

			 [
			  {
			    "Vulnerabilities": [
			      {
			        "Target": ".github/action/package-lock.json",
			        "Type": "npm",
			        "VulnerabilityID": "CVE-2020-15228",
			        "PkgName": "@actions/core",
			        "InstalledVersion": "1.2.4",
				...omitted for brevity...
			        "LastModifiedDate": "2020-12-23T18:32:00Z"
			      },
			      {
			        "Target": "package-lock.json",
			        "Type": "npm",
			        "VulnerabilityID": "CVE-2020-8244",
			        "PkgName": "bl",
			        "InstalledVersion": "4.0.2",
				...omitted for brevity...
			      }

			    ]
			  }
			 ]
	*/

	merged := jnode.NewObjectNode()
	merged.PutArray("Vulnerabilities")
	for i := 0; i < n.Size(); i++ {
		vulns := n.Get(i).Path("Vulnerabilities").Elements()
		for x := range vulns {
			vulns[x].Put("Target", n.Get(i).Path("Target"))
			vulns[x].Put("Type", n.Get(i).Path("Type"))
		}

		err := merged.Path("Vulnerabilities").AppendE(vulns)
		if err != nil {
			log.Errorf("error merging trivy vulns: %w", err)
			break
		}
	}
	return &tools.Result{
		Data: merged,
		Values: map[string]string{
			"TRIVY_VERSION": d.Version,
		},
		PrintPath:    []string{"Vulnerabilities"},
		PrintColumns: []string{"Target", "PkgName", "VulnerabilityID", "Severity", "InstalledVersion", "FixedVersion", "Title"},
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
