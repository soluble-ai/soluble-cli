package tools

import (
	"bytes"
	"io"
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/archive"
	"github.com/soluble-ai/soluble-cli/pkg/client"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/soluble-ai/soluble-cli/pkg/version"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type ToolOpts struct {
	options.PrintClientOpts
	UploadEnabled bool
	OmitContext   bool
	ExitCode      int
}

var _ options.Interface = &ToolOpts{}

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

func (o *ToolOpts) RunTool(tool Interface) error {
	result, err := tool.Run()
	if err != nil {
		return err
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
		err = o.reportResult(tool, result)
		if err != nil {
			return err
		}
	}
	o.Path = result.PrintPath
	o.Columns = result.PrintColumns

	o.PrintResult(result.Data)

	if o.ExitCode != 0 {
		output := print.Nav(result.Data, o.Path)
		if len(output.Elements()) > 0 {
			os.Exit(o.ExitCode)
		}
	}
	return nil
}

func (o *ToolOpts) reportResult(tool Interface, result *Result) error {
	rr := bytes.NewReader([]byte(result.Data.String()))
	log.Infof("Uploading results of {primary:%s}", tool.Name())
	options := []client.Option{
		xcp.WithCIEnv, xcp.WithFileFromReader("results_json", "results.json", rr),
	}
	if !o.OmitContext && result.Files != nil {
		tarball, err := o.createTarball(result)
		if err != nil {
			return err
		}
		defer tarball.Close()
		defer os.Remove(tarball.Name())
		options = append(options, xcp.WithFileFromReader("tarball", "context.tar.gz", tarball))
	}
	return o.GetAPIClient().XCPPost(o.GetOrganization(), tool.Name(), nil, result.Values, options...)
}

func (o *ToolOpts) createTarball(result *Result) (afero.File, error) {
	fs := afero.NewOsFs()
	f, err := afero.TempFile(fs, "", "soluble-cli*")
	if err != nil {
		return nil, err
	}
	tar := archive.NewTarballWriter(f)
	err = util.PropagateCloseError(tar, func() error {
		if result.Files != nil {
			for _, file := range result.Files.Values() {
				if err := tar.WriteFile(fs, result.Directory, file); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return nil, err
	}
	// leave the tarball open, but rewind it to the start
	_, err = f.Seek(0, io.SeekStart)
	return f, err
}
