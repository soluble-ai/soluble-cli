package tools

import (
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
)

type DirectoryBasedToolOpts struct {
	ToolOpts
	Directory string

	absDirectory string
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

func (o *DirectoryBasedToolOpts) Register(cmd *cobra.Command) {
	o.ToolOpts.Register(cmd)
	cmd.Flags().StringVarP(&o.Directory, "directory", "d", "", "The directory to run in.")
}
