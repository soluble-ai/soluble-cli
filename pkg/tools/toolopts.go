package tools

import (
	"bytes"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/version"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
	"github.com/spf13/cobra"
)

type ToolOpts struct {
	options.PrintClientOpts
	ReportEnabled bool
}

var _ options.Interface = &ToolOpts{}

func (o *ToolOpts) Register(c *cobra.Command) {
	flags := c.Flags()
	flags.BoolVar(&o.ReportEnabled, "report", false, "Upload report to Soluble")
}

func (o *ToolOpts) SetContextValues(m map[string]string) {}

func (o *ToolOpts) RunTool(tool Interface) error {
	result, err := tool.Run()
	if err != nil {
		return err
	}
	if result.Values == nil {
		result.Values = map[string]string{}
	}
	result.Values["TOOL_NAME"] = tool.Name()
	result.Values["CLI_VERSION"] = version.Version
	if o.ReportEnabled {
		err = o.reportResult(tool, result)
		if err != nil {
			return err
		}
	}
	o.Path = result.PrintPath
	o.Columns = result.PrintColumns
	o.PrintResult(result.Data)
	return nil
}

func (o *ToolOpts) reportResult(tool Interface, result *Result) error {
	rr := bytes.NewReader([]byte(result.Data.String()))
	log.Infof("Uploading results of {primary:%s}", tool.Name())
	return o.GetAPIClient().XCPPost(o.GetOrganization(), tool.Name(), nil, result.Values,
		xcp.WithCIEnv, xcp.WithFileFromReader("results_json", "results.json", rr))
}
