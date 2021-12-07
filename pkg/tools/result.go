// Copyright 2021 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	Findings     assessments.Findings
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

func (r *Result) Report(tool Interface, upload bool) error {
	return r.report(tool.GetToolOptions(), tool.GetDirectoryBasedToolOptions(), tool.Name(), upload)
}

func (r *Result) report(o *ToolOpts, diropts *DirectoryBasedToolOpts, name string, upload bool) error {
	rr := bytes.NewReader([]byte(r.Data.String()))
	if upload {
		log.Infof("Uploading results of {primary:%s}", name)
	}
	options := []api.Option{
		xcp.WithCIEnv(r.Directory), xcp.WithFileFromReader("results_json", "results.json", rr),
	}
	dir, _ := inventory.FindRepoRoot(r.Directory)
	if dir != "" {
		// include various repo files if they exist
		names := &util.StringSet{}
		for _, path := range repoFiles {
			p := filepath.Join(dir, filepath.FromSlash(path))
			fi, err := os.Stat(p)
			if err != nil || fi.Size() == 0 {
				// don't include 0 length files
				continue
			}
			if f, err := os.Open(p); err == nil {
				defer f.Close()
				name := filepath.Base(path)
				if names.Add(name) {
					// only include one
					options = append(options, xcp.WithFileFromReader(name, name, f))
				}
			}
		}
	}
	if r.Findings != nil {
		if rf := attachFindings(r.Findings); rf != nil {
			options = append(options, xcp.WithFileFromReader("findings_json", "findings.json", rf))
		}
		if rf := attachFingerprints(diropts, r.Findings); rf != nil {
			options = append(options, xcp.WithFileFromReader("fingerprints_json", "fingerprints.json", rf))
		}
	}
	if !upload {
		return nil
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

func attachFindings(findings assessments.Findings) io.Reader {
	fd, err := json.Marshal(findings)
	if err != nil {
		log.Warnf("Could not marshal findings: {warning:%s}", err)
		return nil
	}
	return bytes.NewReader(fd)
}

func attachFingerprints(diropts *DirectoryBasedToolOpts, findings assessments.Findings) io.Reader {
	m := map[string]*assessments.Finding{}
	for _, f := range findings {
		if f.PartialFingerprint == "" {
			continue
		}
		key := fmt.Sprintf("%s:%d", f.FilePath, f.Line)
		ff := m[key]
		if ff == nil {
			m[key] = f
		}
	}
	n := jnode.NewArrayNode()
	for _, f := range m {
		o := n.AppendObject().
			Put("filePath", f.FilePath).
			Put("partialFingerprint", f.PartialFingerprint).
			Put("line", f.Line)
		if f.RepoPath != "" {
			o.Put("repoPath", f.RepoPath)
		}
	}
	if diropts != nil {
		if diropts.PrintFingerprints {
			p := &print.JSONPrinter{}
			p.PrintResult(os.Stderr, n)
		}
		if diropts.SaveFingerprints != "" {
			p := &print.JSONPrinter{}
			f, err := os.Create(diropts.SaveFingerprints)
			if err != nil {
				log.Warnf("Could not save fingerprints: {warning:%s}", err)
			} else {
				p.PrintResult(f, n)
				_ = f.Close()
			}
		}
	}
	d, err := json.Marshal(n)
	if err != nil {
		log.Warnf("Could not marshal fingerprints: {warning:%s}", err)
	}
	return bytes.NewReader(d)
}
