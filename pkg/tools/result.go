package tools

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/archive"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/exit"
	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
	"github.com/spf13/afero"
)

type Result struct {
	Data         *jnode.Node
	PrintData    *jnode.Node
	Values       map[string]string
	Directory    string
	Files        *util.StringSet
	PrintPath    []string
	PrintColumns []string

	Assessment *assessments.Assessment
}

var repoFiles = []string{
	".soluble/config.yml",
	"CODEOWNERS",
	"docs/CODEOWNERS",
	".github/CODEOWNERS",
}

func (r *Result) GetPrintData() *jnode.Node {
	if r.PrintData != nil {
		return r.PrintData
	}
	return r.Data
}

func (r *Result) AddFile(path string) *Result {
	if r.Files == nil {
		r.Files = util.NewStringSet()
	}
	r.Files.Add(path)
	return r
}

func (r *Result) AddValue(name, value string) *Result {
	if r.Values == nil {
		r.Values = map[string]string{}
	}
	r.Values[name] = value
	return r
}

func (r *Result) Report(tool Interface) error {
	return r.report(tool.GetToolOptions(), tool.Name())
}

func (r *Result) report(o *ToolOpts, name string) error {
	rr := bytes.NewReader([]byte(r.Data.String()))
	log.Infof("Uploading results of {primary:%s}", name)
	options := []api.Option{
		xcp.WithCIEnv(r.Directory), xcp.WithFileFromReader("results_json", "results.json", rr),
	}
	if !o.OmitContext {
		if r.Files != nil {
			tarball, err := r.createTarball()
			if err != nil {
				return err
			}
			defer tarball.Close()
			defer os.Remove(tarball.Name())
			options = append(options, xcp.WithFileFromReader("tarball", "context.tar.gz", tarball))
		}
		// include various repo files if they exist
		dir, _ := inventory.FindRepoRoot(r.Directory)
		if dir != "" {
			names := &util.StringSet{}
			for _, path := range repoFiles {
				if f, err := os.Open(filepath.Join(dir, filepath.FromSlash(path))); err == nil {
					defer f.Close()
					name := filepath.Base(path)
					if names.Add(name) {
						// only include one
						options = append(options, xcp.WithFileFromReader(name, name, f))
					}
				}
			}
		}
	}
	n, err := o.GetAPIClient().XCPPost(o.GetOrganization(), name, nil, r.Values, options...)
	if err != nil {
		return err
	}
	if o.PrintAsessment {
		p := &print.YAMLPrinter{}
		p.PrintResult(os.Stderr, n)
	}
	if n.Path("assessment").IsObject() {
		r.Assessment = &assessments.Assessment{}
		if err := json.Unmarshal([]byte(n.Path("assessment").String()), r.Assessment); err != nil {
			log.Warnf("The server returned a garbled assessment: {warning:%s}", err)
			r.Assessment = nil
		}
	}
	if r.Assessment == nil {
		log.Infof("No assessment was returned")
	} else {
		r.Assessment.EvaluateFailures(o.ParsedFailThresholds)
		if r.Assessment.Failed {
			exit.Code = 2
			exit.AddFunc(func() {
				log.Errorf("Found {danger:%d failed %s} findings",
					r.Assessment.FailedCount, r.Assessment.FailedSeverity)
			})
		}
	}
	return nil
}

func (r *Result) createTarball() (afero.File, error) {
	fs := afero.NewOsFs()
	f, err := afero.TempFile(fs, "", "soluble-cli*")
	if err != nil {
		return nil, err
	}
	tar := archive.NewTarballWriter(f)
	err = util.PropagateCloseError(tar, func() error {
		if r.Files != nil {
			for _, file := range r.Files.Values() {
				if err := tar.WriteFile(fs, r.Directory, file); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return nil, err
	}
	// leave the tarball open, but rewind it to the start
	_, err = f.Seek(0, io.SeekStart)
	return f, err
}
