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
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type DirectoryBasedToolOpts struct {
	ToolOpts
	Directory         string
	Exclude           []string
	PrintFingerprints bool
	SaveFingerprints  string

	absDirectory string
	ignore       *ignore.GitIgnore
}

func (o *DirectoryBasedToolOpts) GetDirectoryBasedToolOptions() *DirectoryBasedToolOpts {
	return o
}

func (o *DirectoryBasedToolOpts) GetDirectory() string {
	if o.absDirectory == "" {
		dir := o.Directory
		if dir == "" {
			dir = "."
		}
		dir, err := filepath.Abs(dir)
		if err != nil {
			log.Errorf("Cannot determine current directory: {danger:%s}", err)
			panic(err)
		}
		o.absDirectory = dir
	}
	return o.absDirectory
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

func (o *DirectoryBasedToolOpts) GetDirectoryBasedHiddenOptions() *options.HiddenOptionsGroup {
	return &options.HiddenOptionsGroup{
		Name: "directory-based-options",
		Long: "Options for running tools in a directory",
		CreateFlagsFunc: func(flags *pflag.FlagSet) {
			flags.BoolVar(&o.PrintFingerprints, "print-fingerprints", false, "Print fingerprints on stderr before uploading results")
			flags.StringVar(&o.SaveFingerprints, "save-fingerprints", "", "Save finding fingerprints to `file`")
		},
	}
}

func (o *DirectoryBasedToolOpts) Register(cmd *cobra.Command) {
	o.ToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVarP(&o.Directory, "directory", "d", "", "The directory to run in.")
	flags.StringSliceVar(&o.Exclude, "exclude", nil, "Exclude results from file that match this glob pattern (path/**/foo.txt syntax supported.)  May be repeated.")
	o.GetDirectoryBasedHiddenOptions().Register(cmd)
}

func (o *DirectoryBasedToolOpts) Validate() error {
	o.absDirectory = ""
	if o.RepoRoot == "" {
		var err error
		o.RepoRoot, err = inventory.FindRepoRoot(o.GetDirectory())
		if err != nil {
			return err
		}
	}
	if err := o.ToolOpts.Validate(); err != nil {
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
