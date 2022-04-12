package tools

import (
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/repotree"
	"github.com/spf13/cobra"
)

type DirectoryOpt struct {
	Directory    string
	absDirectory string
}

type HasDirectory interface {
	GetDirectory() string
	SetDirectory(dir string)
}

func (o *DirectoryOpt) GetDirectory() string {
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

func (o *DirectoryOpt) SetDirectory(dir string) {
	o.absDirectory = ""
	o.Directory = dir
}

func (o *DirectoryOpt) Register(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(&o.Directory, "directory", "d", "", "The directory to run in.")
}

func (o *DirectoryOpt) Validate(opts *ToolOpts) error {
	o.absDirectory = ""
	if opts.RepoRoot == "" {
		var err error
		opts.RepoRoot, err = repotree.FindRepoRoot(o.GetDirectory())
		if err != nil {
			return err
		}
		opts.repoRootSet = true
	}
	return nil
}
