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

package tools

import (
	"fmt"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/spf13/cobra"
)

type HasCommandTemplate interface {
	CommandTemplate() *cobra.Command
}

func CreateCommand(tool Interface) *cobra.Command {
	var c *cobra.Command
	if ct, ok := tool.(HasCommandTemplate); ok {
		c = ct.CommandTemplate()
		if c.Args == nil {
			c.Args = cobra.NoArgs
		}
	} else {
		c = &cobra.Command{
			Use:   tool.Name(),
			Short: fmt.Sprintf("Run %s", tool.Name()),
			Args:  cobra.NoArgs,
		}
	}
	c.RunE = func(cmd *cobra.Command, args []string) error {
		return runTool(tool)
	}
	tool.Register(c)
	return c
}

func runTool(tool Interface) error {
	opts := tool.GetToolOptions()
	opts.Tool = tool
	var (
		results Results
		toolErr error
	)
	switch t := tool.(type) {
	case Simple:
		if err := t.Validate(); err != nil {
			return err
		}
		return t.Run()
	case Single:
		var r *Result
		r, toolErr = RunSingleAssessment(t)
		if r != nil {
			results = Results{r}
		}
	case Consolidated:
		results, toolErr = RunConsoliatedAssessments(t)
	default:
		panic("tools must implement Simple, Single or Conslidated")
	}
	// even if the tool had an error we may have partial
	// results that can be displayed
	for _, result := range results {
		if result.Assessment != nil && result.Assessment.URL != "" {
			log.Infof("Asessment uploaded, see {primary:%s} for more information", result.Assessment.URL)
		}
	}
	var (
		n   *jnode.Node
		err error
	)
	// What we really want to work off here is a list of all the assessments.
	// But the printer doesn't support a splat-like path i.e. *.findings to
	// accumulate all the findings across the assessments.  So for the default
	// or table output format we do that accumulation in code here.

	switch {
	case opts.OutputFormat == "" || opts.OutputFormat == "table" || opts.OutputFormat == "count":
		if !opts.Wide {
			opts.SetFormatter("title", print.TruncateFormatter(70, false))
			opts.SetFormatter("filePath", print.TruncateFormatter(65, true))
		}
		n, err = results.getFindingsJNode()
	default:
		n, err = results.getAssessmentsJNode()
	}

	if err != nil {
		return err
	}
	if toolErr == nil {
		opts.PrintResult(n)
	}
	if toolErr != nil {
		return toolErr
	}
	return nil
}
