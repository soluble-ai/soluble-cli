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

type DirectoryBasedToolOpts struct {
	ToolOpts
	Directory         string
	Exclude           []string
	PrintFingerprints bool

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

func (o *DirectoryBasedToolOpts) Register(cmd *cobra.Command) {
	o.ToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVarP(&o.Directory, "directory", "d", "", "The directory to run in.")
	flags.StringSliceVar(&o.Exclude, "exclude", nil, "Exclude results from file that match this glob pattern (path/**/foo.txt syntax supported.)  May be repeated.")
	flags.BoolVar(&o.PrintFingerprints, "print-fingerprints", false, "Print fingerprints on stderr before uploading results")
	_ = flags.MarkHidden("print-fingerprints")
}

func (o *DirectoryBasedToolOpts) Validate() error {
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
