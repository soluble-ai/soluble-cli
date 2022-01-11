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
	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
)

type Result struct {
	Data             *jnode.Node
	Findings         assessments.Findings
	Values           map[string]string
	Directory        string
	Files            *util.StringSet
	PrintPath        []string
	PrintColumns     []string
	FileFingerprints []*FileFingerprint

	Assessment *assessments.Assessment
}

type Results []*Result

type FileFingerprint struct {
	Line               int    `json:"line"`
	RepoPath           string `json:"repoPath,omitempty"`
	PartialFingerprint string `json:"partialFingerprint,omitempty"`
	FilePath           string `json:"filePath"`
}

var repoFiles = []string{
	".lacework/config.yml",
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

func (r *Result) Upload(client *api.Client, org, name string) error {
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
		if rf := r.attachFindings(); rf != nil {
			options = append(options, xcp.WithFileFromReader("findings_json", "findings.json", rf))
		}
		if rf := r.attachFingerprints(); rf != nil {
			options = append(options, xcp.WithFileFromReader("fingerprints_json", "fingerprints.json", rf))
		}
	}
	n, err := client.XCPPost(org, name, nil, r.Values, options...)
	if err != nil {
		return err
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
	}
	return nil
}

func (r *Result) UpdateFileFingerprints() {
	if r.Directory == "" {
		return
	}
	r.Findings.ComputePartialFingerprints(r.Directory)
	m := map[string]*assessments.Finding{}
	for _, f := range r.Findings {
		if f.PartialFingerprint == "" {
			continue
		}
		key := fmt.Sprintf("%s:%d", f.FilePath, f.Line)
		ff := m[key]
		if ff == nil {
			m[key] = f
		}
	}
	r.FileFingerprints = make([]*FileFingerprint, 0, len(m))
	for _, f := range m {
		r.FileFingerprints = append(r.FileFingerprints,
			&FileFingerprint{
				FilePath:           f.FilePath,
				PartialFingerprint: f.PartialFingerprint,
				Line:               f.Line,
				RepoPath:           f.RepoPath,
			})
	}
}

func (r *Result) attachFindings() io.Reader {
	fd, err := json.Marshal(r.Findings)
	if err != nil {
		log.Warnf("Could not marshal findings: {warning:%s}", err)
		return nil
	}
	return bytes.NewReader(fd)
}

func (r *Result) attachFingerprints() io.Reader {
	d, err := json.Marshal(r.FileFingerprints)
	if err != nil {
		log.Warnf("Could not marshal fingerprints: {warning:%s}", err)
	}
	return bytes.NewReader(d)
}

func (results Results) getFindingsJNode() (*jnode.Node, error) {
	var findings []*assessments.Finding
	for _, result := range results {
		if result.Assessment != nil {
			findings = append(findings, result.Assessment.Findings...)
		} else {
			findings = append(findings, result.Findings...)
		}
	}
	d, err := json.Marshal(findings)
	if err != nil {
		return nil, err
	}
	return jnode.FromJSON(d)
}
