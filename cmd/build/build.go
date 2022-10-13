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

package build

import (
	"fmt"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/exit"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "build",
		Short: "Commands for CI builds",
	}
	c.AddCommand(
		buildReportCommand(),
		updatePRCommand(),
	)
	return c
}

func buildReportCommand() *cobra.Command {
	opts := &options.PrintClientOpts{
		PrintOpts: options.PrintOpts{
			Path:    []string{"findings"},
			Columns: []string{"module", "pass", "severity", "sid", "file:line", "title"},
		},
	}
	opts.SetFormatter("pass", tools.PassFormatter)
	c := &cobra.Command{
		Use:   "report",
		Short: "Display any assessments generated during this build",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			api, err := opts.GetAPIClient()
			if err != nil {
				return err
			}
			assessments, err := assessments.FindCIEnvAssessments(api)
			if err != nil {
				return err
			}
			findings := jnode.NewArrayNode()
			for _, assessment := range assessments {
				if assessment.Failed {
					exit.Code = 2
					a := assessment
					exit.AddFunc(func() {
						log.Errorf("{warning:%s} has {danger:%d %s findings}",
							a.Title, a.FailedCount, a.FailedSeverity)
					})
				}
				for _, finding := range assessment.Findings {
					findings.AppendObject().Put("sid", finding.SID).
						Put("module", assessment.Module).
						Put("pass", finding.Pass).
						Put("severity", finding.Severity).
						Put("file:line", fmt.Sprintf("%s:%d", finding.FilePath, finding.Line)).
						Put("title", finding.GetTitle())
				}
			}
			opts.PrintResult(jnode.NewObjectNode().Put("findings", findings))
			for _, assessment := range assessments {
				if assessment.URL != "" {
					log.Infof("For more details on the {info:%s} see {primary:%s}", assessment.Title, assessment.URL)
				}
			}
			return nil
		},
	}
	opts.Register(c)
	return c
}

func updatePRCommand() *cobra.Command {
	opts := &options.PrintClientOpts{}
	c := &cobra.Command{
		Use:    "update-pr",
		Short:  "Update this build's pull-request with the results of any assessments generated during the build",
		Args:   cobra.NoArgs,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			api, err := opts.GetAPIClient()
			if err != nil {
				return err
			}
			assessments, err := assessments.FindCIEnvAssessments(api)
			if err != nil {
				return err
			}
			return assessments.UpdatePR(api)
		},
	}
	opts.Register(c)
	return c
}
