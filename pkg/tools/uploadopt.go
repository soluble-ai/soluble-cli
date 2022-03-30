package tools

import (
	"github.com/spf13/cobra"
)

type UploadOpt struct {
	UploadEnabled bool
}

func (o *UploadOpt) Register(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.BoolVar(&o.UploadEnabled, "upload", false, "Upload results")
}
