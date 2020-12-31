package buildreport

import (
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/exit"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

type BuildReportOpts struct {
	options.PrintClientOpts
	FailThresholds map[string]string

	parsedFailThresholds map[string]int
}

func (opts *BuildReportOpts) Register(c *cobra.Command) {
	opts.PrintClientOpts.Register(c)
	flags := c.Flags()
	flags.StringToStringVar(&opts.FailThresholds, "fail", nil, "")
}

func (opts *BuildReportOpts) getAssessments() (assessments.Assessments, error) {
	parsedFailThresholds, err := assessments.ParseFailThresholds(opts.FailThresholds)
	if err != nil {
		return nil, err
	}
	opts.parsedFailThresholds = parsedFailThresholds
	return assessments.FindCIEnvAssessments(opts.GetAPIClient())
}

func Commands() []*cobra.Command {
	return []*cobra.Command{
		buildReportCommand(),
		updateCICommand(),
	}
}

func buildReportCommand() *cobra.Command {
	opts := &BuildReportOpts{
		PrintClientOpts: options.PrintClientOpts{
			PrintOpts: options.PrintOpts{
				Path:    []string{"findings"},
				Columns: []string{"module", "pass", "severity", "sid", "description"},
			},
		},
	}
	c := &cobra.Command{
		Use:   "build-report",
		Short: "List any assessments generated during this build",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			assessments, err := opts.getAssessments()
			if err != nil {
				return err
			}
			findings := jnode.NewArrayNode()
			for _, assessment := range assessments {
				failed, count, level := assessment.HasFailures(opts.parsedFailThresholds)
				if failed {
					exit.Code = 2
					exit.AddFunc(func() { log.Errorf("{warning:%s} has {danger:%d %s findings}", assessment.Title, count, level) })
				}
				for _, finding := range assessment.Findings {
					findings.AppendObject().Put("sid", finding.SID).
						Put("module", assessment.Module).
						Put("pass", finding.Pass).
						Put("severity", finding.Severity).
						Put("file", finding.FilePath).
						Put("title", finding.GetTitle())
				}
			}
			opts.PrintResult(jnode.NewObjectNode().Put("findings", findings))
			return nil
		},
	}
	opts.Register(c)
	c.Flag("fail").Usage = longUsage(`
Set failure thresholds in the form 'severity=count'.  The command will exit with exit code 2 
if the assessments generated during this build have count or more failed findings of the
specified severity.`)
	return c
}

func updateCICommand() *cobra.Command {
	opts := &BuildReportOpts{}
	c := &cobra.Command{
		Use:   "update-ci",
		Short: "Update the CI system with the results of any assessments generated during the build",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			assessments, err := opts.getAssessments()
			if err != nil {
				return err
			}
			return assessments.UpdateCI(opts.GetAPIClient())
		},
	}
	opts.Register(c)
	c.Flag("fail").Usage = longUsage(`
Set failure thresholds in the form 'severity=count'.  The checks that this command creates
will be marked as failed if the corresponding assessment has count or more failed findings
of the specified severity.`)
	return c
}

func longUsage(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(strings.ReplaceAll(s, "'", "`")), "\n", " ")
}
