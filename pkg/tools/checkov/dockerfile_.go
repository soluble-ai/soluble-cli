package checkov

// NB - naming this file "dockerfile.go" breaks visual studio code

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
		df := d.Dockerfile
		if !filepath.IsAbs(df) {
			var err error
			df, err = filepath.Abs(df)
			if err != nil {
				return err
			}
		}
		if !util.FileExists(df) {
			return fmt.Errorf("%s not found", df)
		}
		if d.Directory != "" {
			log.Warnf("Using --dockerfile will override the use of --directory")
		}
		d.Directory = filepath.Dir(df)
		d.dockerfile = df
	} else {
		if err := d.DirectoryBasedToolOpts.Validate(); err != nil {
			return err
		}
		df := filepath.Join(d.GetDirectory(), "Dockerfile")
		if !util.FileExists(df) {
			df = filepath.Join(d.GetDirectory(), "dockerfile")
			if !util.FileExists(df) {
				return fmt.Errorf("neither Dockerfile nor dockerfile exist in %s", d.GetDirectory())
			}
		}
		d.dockerfile = df
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
