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
	"fmt"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

// Options for assessment tools that run in a directory
type DirectoryBasedToolOpts struct {
	AssessmentOpts
	DirectoryOpt
	Exclude []string
	ignore  *ignore.GitIgnore
}

func (o *DirectoryBasedToolOpts) GetInventory() *inventory.Manifest {
	m := inventory.Do(o.GetDirectory())
	m.CloudformationFiles = o.removeExcludedStringSet(m.CloudformationFiles)
	m.DockerDirectories = o.removeExcludedStringSet(m.DockerDirectories)
	m.HelmCharts = o.removeExcludedStringSet(m.HelmCharts)
	m.KubernetesManifestDirectories = o.removeExcludedStringSet(m.KubernetesManifestDirectories)
	m.TerraformRootModules = o.removeExcludedStringSet(m.TerraformRootModules)
	m.TerraformModules = o.removeExcludedStringSet(m.TerraformModules)
	return m
}

func (o *DirectoryBasedToolOpts) GetFilesInDirectory(files []string) ([]string, error) {
	var result []string
	for _, f := range files {
		if filepath.IsAbs(f) {
			if !strings.HasPrefix(f, o.GetDirectory()) {
				return nil, fmt.Errorf("file %s must be relative to --directory", f)
			}
		}
		if !o.IsExcluded(f) {
			result = append(result, f)
		}
	}
	return result, nil
}

func (o *DirectoryBasedToolOpts) RemoveExcluded(files []string) []string {
	var result []string
	for _, f := range files {
		if !o.IsExcluded(f) {
			result = append(result, f)
		}
	}
	return result
}

func (o *DirectoryBasedToolOpts) removeExcludedStringSet(ss util.StringSet) util.StringSet {
	var r util.StringSet
	for _, v := range ss.Values() {
		if !o.IsExcluded(v) {
			r.Add(v)
		}
	}
	return r
}

func (o *DirectoryBasedToolOpts) IsExcluded(file string) bool {
	if o.ignore != nil {
		rfile := MustRel(o.GetDirectory(), file)
		if o.ignore.MatchesPath(rfile) {
			return true
		}
	}
	if o.RepoRoot == "" {
		return false
	}
	rfile := MustRel(o.RepoRoot, file)
	return o.GetConfig().IsIgnored(rfile)
}

// Return the directory that a docker-based tool is run in.  Normally
// this is /src, but if it's run out of PATH, then it's o.GetDirectory()
func (o *DirectoryBasedToolOpts) GetDockerRunDirectory() string {
	if o.ToolPath != "" || o.NoDocker {
		return o.GetDirectory()
	}
	return "/src"
}

func (o *DirectoryBasedToolOpts) Register(cmd *cobra.Command) {
	o.AssessmentOpts.Register(cmd)
	o.DirectoryOpt.Register(cmd)
	flags := cmd.Flags()
	flags.StringSliceVar(&o.Exclude, "exclude", nil, "Exclude results from file that match this glob `pattern` (path/**/foo.txt syntax supported.)  May be repeated.")
}

func (o *DirectoryBasedToolOpts) Validate() error {
	if err := o.DirectoryOpt.Validate(&o.ToolOpts); err != nil {
		return err
	}
	if err := o.AssessmentOpts.Validate(); err != nil {
		return err
	}
	if len(o.Exclude) > 0 {
		o.ignore = ignore.CompileIgnoreLines(o.Exclude...)
		if o.ignore == nil {
			log.Warnf("Invalid exclude pattern {warning:%s}", strings.Join(o.Exclude, ","))
		}
	}
	return nil
}
