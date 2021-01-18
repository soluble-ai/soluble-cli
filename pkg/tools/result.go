package tools

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/exit"
	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
)

type Result struct {
	Data         *jnode.Node
	Findings     []*assessments.Finding
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
	dir, _ := inventory.FindRepoRoot(r.Directory)
	if dir != "" {
		// include various repo files if they exist
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
		if r.Findings != nil {
			fd, err := json.Marshal(r.Findings)
			if err != nil {
				log.Warnf("Could not marshal findings: {warning:%s}", err)
			} else {
				rf := bytes.NewReader(fd)
				options = append(options, xcp.WithFileFromReader("findings_json", "findings.json", rf))
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
		log.Infof("No assessment for {warning:%s} was returned", name)
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
