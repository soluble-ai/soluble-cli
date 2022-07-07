package checkov

import (
	"fmt"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Dockerfile struct {
	tools.DirectoryBasedToolOpts
	Dockerfile string

	dockerfile string
}

var _ tools.Interface = (*Dockerfile)(nil)

func (*Dockerfile) Name() string {
	return "checkov-docker"
}

func (d *Dockerfile) Register(cmd *cobra.Command) {
	d.DirectoryBasedToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVar(&d.Dockerfile, "dockerfile", "", "Scan `dockerfile` explicitly.  By default, look for Dockerfile in the target directory.")
}

func (d *Dockerfile) Validate() error {
	if d.Dockerfile != "" {
		if d.Directory != "" {
			log.Warnf("Using --dockerfile will override the use of --directory")
		}
		d.Directory = filepath.Dir(d.Dockerfile)
		d.dockerfile = "Dockerfile"
	}
	if err := d.DirectoryBasedToolOpts.Validate(); err != nil {
		return err
	}
	if d.dockerfile == "" {
		d.dockerfile = filepath.Join(d.GetDirectory(), "Dockerfile")
	}
	if !util.FileExists(d.dockerfile) {
		return fmt.Errorf("%s not found", d.dockerfile)
	}
	return nil
}

func (d *Dockerfile) Run() (*tools.Result, error) {
	checkov := &Tool{
		DirectoryBasedToolOpts: d.DirectoryBasedToolOpts,
		Framework:              "dockerfile",
		targetFile:             d.dockerfile,
	}
	if err := checkov.Validate(); err != nil {
		return nil, err
	}
	return checkov.Run()
}
