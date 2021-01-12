package tools

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v2"
	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type DirectoryBasedToolOpts struct {
	ToolOpts
	Directory string
	Exclude   []string

	absDirectory string
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
			log.Fatalf("Cannot determine current directory: {danger:%s}", err)
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
	m.TerraformRootModuleDirectories = o.removeExcludedStringSet(m.TerraformRootModuleDirectories)
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
	if filepath.IsAbs(file) {
		file, _ = filepath.Abs(file)
	}
	for _, pat := range o.Exclude {
		m, _ := doublestar.PathMatch(pat, file)
		if m {
			return true
		}
	}
	return false
}

func (o *DirectoryBasedToolOpts) Register(cmd *cobra.Command) {
	o.ToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVarP(&o.Directory, "directory", "d", "", "The directory to run in.")
	flags.StringSliceVar(&o.Exclude, "exclude", nil, "Exclude results from file that match this glob pattern (path/**/foo.txt syntax supported.)  May be repeated.")
}

func (o *DirectoryBasedToolOpts) Validate() error {
	if err := o.ToolOpts.Validate(); err != nil {
		return err
	}
	for _, pat := range o.Exclude {
		_, err := doublestar.PathMatch(pat, "test")
		if err != nil {
			return fmt.Errorf("invalid --exclude pattern '%s': %w", pat, err)
		}
	}
	return nil
}
