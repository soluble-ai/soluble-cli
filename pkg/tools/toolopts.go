package tools

import (
	"github.com/soluble-ai/soluble-cli/pkg/client"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

type ToolOpts struct {
	options.PrintClientOpts
	UploadEnabled bool
	OmitContext   bool
}

var _ options.Interface = &ToolOpts{}

func (o *ToolOpts) GetToolOptions() *ToolOpts {
	return o
}

func (o *ToolOpts) Register(c *cobra.Command) {
	// set this now so help shows up, it will be corrected before we print anything
	o.Path = []string{}
	o.AuthNotRequired = true
	o.PrintClientOpts.Register(c)
	flags := c.Flags()
	flags.BoolVar(&o.UploadEnabled, "upload", false, "Upload report to Soluble")
	flags.BoolVar(&o.OmitContext, "omit-context", false, "Don't include the source files with violations in the upload")
}

func (o *ToolOpts) SetContextValues(m map[string]string) {}

func (o *ToolOpts) InstallAPIServerArtifact(name, urlPath string) (*download.Download, error) {
	apiClient := o.GetAPIClient()
	m := download.NewManager()
	return m.Install(&download.Spec{
		Name:              name,
		APIServerArtifact: urlPath,
		APIServer:         apiClient.(*client.Client),
	})
}
