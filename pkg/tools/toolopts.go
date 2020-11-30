package tools

import (
	"fmt"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/client"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
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

func (o *ToolOpts) RunDocker(d *DockerTool) ([]byte, error) {
	n := o.getToolVersion(d.Name)
	if image := n.Path("image"); !image.IsMissing() {
		d.Image = image.AsText()
	}
	return d.run()
}

func (o *ToolOpts) InstallTool(spec *download.Spec) (*download.Download, error) {
	if strings.HasPrefix(spec.URL, "github.com/") {
		slash := strings.LastIndex(spec.URL, "/")
		n := o.getToolVersion(spec.URL[slash+1:])
		if v := n.Path("version"); !v.IsMissing() {
			spec.RequestedVersion = v.AsText()
		}
	}
	m := download.NewManager()
	return m.Install(spec)
}

func (o *ToolOpts) getToolVersion(name string) *jnode.Node {
	temp := log.SetTempLevel(log.Warning)
	defer temp.Restore()
	n, err := o.GetUnauthenticatedAPIClient().Get(fmt.Sprintf("cli/tools/%s/config", name))
	if err != nil {
		log.Warnf("Could not get version of {primary:%s}: {warning:%s}", name, err)
		return jnode.MissingNode
	}
	return n
}
