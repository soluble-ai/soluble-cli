// Copyright 2021 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package autoscan

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	cfnpythonlint "github.com/soluble-ai/soluble-cli/pkg/tools/cfn-python-lint"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/soluble-ai/soluble-cli/pkg/tools/iacinventory"
	"github.com/soluble-ai/soluble-cli/pkg/tools/secrets"
	"github.com/soluble-ai/soluble-cli/pkg/tools/trivy"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
	PrintToolResults bool
	Skip             []string
	ToolPaths        map[string]string
	Images           []string
}

var _ tools.Interface = &Tool{}

type SubordinateTool struct {
	tools.Interface
	Skip   bool
	Result *tools.Result
	Err    error
}

func (*Tool) Name() string {
	return "autoscan"
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.Internal = true
	t.DirectoryBasedToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.BoolVar(&t.PrintToolResults, "print-tool-results", false, "Print individual results from tools")
	flags.StringSliceVar(&t.Skip, "skip", nil, "Don't run these `tools` (command-separated or repeated.)")
	flags.StringToStringVar(&t.ToolPaths, "tool-paths", nil, "Explicitly specify the path to each tool in the form `tool=path`.")
	flags.StringSliceVar(&t.Images, "image", nil, "Scan these docker images, as in the image-scan command.")
	flags.BoolVar(&t.NoDocker, "no-docker", false, "Run all docker-based tools locally")
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "auto-scan",
		Short: "Find infrastructure-as-code and scan with recommended tools",
		Long: `Find infrastructure-as-code and scan with the following tools:

Cloudformation templates - cfn-python-lint
Terraform                - checkov
Kuberentes manifests     - checkov
Everything               - secrets		

In addition, images can be scanned with trivy.
`,
		Example: `# To run a tool locally w/o using docker explicitly specify the tool path
... auto-scan --tool-paths checkov=checkov,cfn-python-lint=cfn-lint`,
		Hidden: true,
	}
}

func (t *Tool) Run() (*tools.Result, error) {
	m := inventory.Do(t.GetDirectory())
	subTools := []SubordinateTool{
		{
			Interface: &iacinventory.Local{
				DirectoryBasedToolOpts: t.getDirectoryOpts(),
			},
		},
		{
			Interface: &checkov.Tool{
				DirectoryBasedToolOpts: t.getDirectoryOpts(),
			},
			Skip: m.TerraformRootModules.Len() == 0 && m.KubernetesManifestDirectories.Len() == 0,
		},
		{
			Interface: &cfnpythonlint.Tool{
				DirectoryBasedToolOpts: t.getDirectoryOpts(),
				Templates:              m.CloudformationFiles.Values(),
			},
			Skip: m.CloudformationFiles.Len() == 0,
		},
		{
			Interface: &secrets.Tool{
				DirectoryBasedToolOpts: t.getDirectoryOpts(),
			},
		},
	}
	for _, image := range t.Images {
		subTools = append(subTools, SubordinateTool{
			Interface: &trivy.Tool{
				Image: image,
			},
		})
	}
	result := &tools.Result{
		Data:      jnode.NewObjectNode(),
		PrintPath: []string{"data"},
		PrintColumns: []string{
			"name", "run_duration", "findings_count", "error",
			"assessment_url"},
	}
	resultData := result.Data.PutArray("data")
	count := 0
	var errs error
	for _, st := range subTools {
		n := resultData.AppendObject()
		n.Put("skipped", st.Skip)
		n.Put("name", st.Name())
		if st.Skip || util.StringSliceContains(t.Skip, st.Name()) {
			n.Put("run_duration", "skipped")
			continue
		}
		count++
		opts := st.GetToolOptions()
		opts.Tool = st
		opts.UploadEnabled = t.UploadEnabled
		opts.ToolPath = t.ToolPaths[st.Name()]
		opts.NoDocker = t.NoDocker
		if dopts := st.GetDirectoryBasedToolOptions(); dopts != nil {
			dopts.Exclude = t.Exclude
		}
		start := time.Now()
		st.Result, st.Err = opts.RunTool(false)
		rd := time.Since(start).Truncate(time.Millisecond)
		n.Put("run_duration", rd.String())
		if st.Result != nil {
			opts.Path = st.Result.PrintPath
			opts.Columns = st.Result.PrintColumns
			if t.PrintToolResults {
				opts.PrintToolResult(st.Result)
			}
			if pr, err := st.GetToolOptions().GetPrinter(); err == nil {
				if st.Result.Findings != nil {
					n.Put("findings_count", len(st.Result.Findings))
				} else if tp, ok := pr.(*print.TablePrinter); ok {
					n.Put("findings_count", len(tp.GetRows(st.Result.Data)))
				}
			}
			if st.Result.Assessment != nil && st.Result.Assessment.URL != "" {
				n.Put("assessment_url", st.Result.Assessment.URL)
			}
		}
		if st.Err != nil {
			n.Put("error", st.Err.Error())
			errs = multierror.Append(errs, fmt.Errorf("%s failed - %w", st.Name(), st.Err))
		}
	}
	log.Infof("Finished running {primary:%d} tools", count)
	return result, errs
}

func (t *Tool) getDirectoryOpts() tools.DirectoryBasedToolOpts {
	return tools.DirectoryBasedToolOpts{
		Directory: t.GetDirectory(),
	}
}
