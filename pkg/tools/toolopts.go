package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/client"
	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/version"
	"github.com/spf13/cobra"
)

type ToolOpts struct {
	options.PrintClientOpts
	UploadEnabled bool
	OmitContext   bool
	ToolVersion   string
	ToolPath      string
	Internal      bool
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
	flags.BoolVar(&o.OmitContext, "omit-context", false, "When uploading a report, don't include the source files with findings")
	if !o.Internal {
		flags.StringVar(&o.ToolPath, "tool-path", "", "Run `tool` directly instead of using a CLI-managed version")
		flags.StringVar(&o.ToolVersion, "tool-version", "", "Override version of the tool to run (the image or github release name.)")
	}
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
	if o.ToolPath != "" {
		// don't use docker, just run it directly
		// #nosec G204
		c := exec.Command(o.ToolPath, d.Args...)
		c.Dir = d.Directory
		c.Stderr = os.Stderr
		log.Infof("Running {primary:%s} {secondary:(in %s)}", strings.Join(c.Args, " "), c.Dir)
		return c.Output()
	}
	n := o.getToolVersion(d.Name)
	if image := n.Path("image"); !image.IsMissing() {
		d.Image = image.AsText()
	}
	return d.run()
}

func (o *ToolOpts) InstallTool(spec *download.Spec) (*download.Download, error) {
	if o.ToolPath != "" {
		return &download.Download{
			OverrideExe: o.ToolPath,
		}, nil
	}
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
	if o.ToolVersion != "" {
		return jnode.NewObjectNode().
			Put("image", o.ToolVersion).
			Put("version", o.ToolVersion)
	}
	temp := log.SetTempLevel(log.Error - 1)
	defer temp.Restore()
	n, err := o.GetUnauthenticatedAPIClient().Get(fmt.Sprintf("cli/tools/%s/config", name))
	if err != nil {
		return jnode.MissingNode
	}
	return n
}

func (o *ToolOpts) RunTool(tool Interface) (*Result, error) {
	result, err := tool.Run()
	if err != nil || result == nil {
		return nil, err
	}
	result.AddValue("TOOL_NAME", tool.Name()).
		AddValue("CLI_VERSION", version.Version)
	if result.Data != nil && result.PrintPath != nil {
		// include the print config in the results
		p := result.Data.PutObject("soluble_print_config")
		p.Put("print_path", jnode.FromSlice(result.PrintPath))
		p.Put("print_columns", jnode.FromSlice(result.PrintColumns))
	}
	if o.UploadEnabled {
		err = result.Report(tool)
	}
	return result, err
}
