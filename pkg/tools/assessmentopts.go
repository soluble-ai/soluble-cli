package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/archive"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/spf13/afero"

	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type AssessmentOpts struct {
	ToolOpts
	UploadOpts
	PrintResultOpt            bool
	SaveResult                string
	PrintResultValues         bool
	SaveResultValues          string
	DisableCustomPolicies     bool
	PrintFingerprints         bool
	SaveFingerprints          string
	CustomPoliciesDir         string
	PreparedCustomPoliciesDir string
	FailThresholds            []string

	parsedFailThresholds   map[string]int
	CustomPolicyMetadata   map[string]string
	LaceworkPolicyMetadata map[string]string
}

func (o *AssessmentOpts) GetAssessmentOptions() *AssessmentOpts {
	return o
}

func (o *AssessmentOpts) Register(c *cobra.Command) {
	o.ToolOpts.Register(c)
	o.DefaultUploadEnabled = true
	o.UploadOpts.Register(c)
	o.SetFormatter("pass", PassFormatter)
	// if not uploaded these columns will be empty, so make that a little easier to see
	o.SetFormatter("sid", MissingFormatter)
	o.SetFormatter("severity", MissingFormatter)
	o.GetAssessmentHiddenOptions().Register(c)
	o.Path = []string{}
	o.Columns = []string{
		"sid", "severity", "pass", "title", "filePath", "line",
	}
}

func (o *AssessmentOpts) GetAssessmentHiddenOptions() *options.HiddenOptionsGroup {
	return &options.HiddenOptionsGroup{
		Name: "tool-options",
		Long: "Options for running tools",
		Example: config.ExpandCommandInvocation(`
A tool run can optionally exit with exit code 2 if the assessment contains
failed findings.  For example:
		
# Fail if 1 or more high or critical severity findings in this build:
{{ .CommandInvocation }} ... --fail high=1
# Or shorter:
{{ .CommandInvocation }} ... --fail high

The severity levels are critical, high, medium, low, and info in that order.`),
		CreateFlagsFunc: func(flags *pflag.FlagSet) {
			flags.BoolVar(&o.DisableCustomPolicies, "disable-custom-policies", false, "Don't use custom policies")
			flags.StringVar(&o.CustomPoliciesDir, "custom-policies", "", "Use opal custom policies from `dir` to run a local assessment, setting upload=false for assessments.")
			flags.BoolVar(&o.PrintResultOpt, "print-result", false, "Print the JSON result from the tool on stderr")
			flags.StringVar(&o.SaveResult, "save-result", "", "Save the JSON result from the tool to `file`")
			flags.BoolVar(&o.PrintResultValues, "print-result-values", false, "Print the result values from the tool on stderr")
			flags.StringVar(&o.SaveResultValues, "save-result-values", "", "Save the result values from the tool to `file`")
			flags.BoolVar(&o.PrintFingerprints, "print-fingerprints", false, "Print fingerprints on stderr before uploading results")
			flags.StringVar(&o.SaveFingerprints, "save-fingerprints", "", "Save finding fingerprints to `file`")
			flags.StringSliceVar(&o.FailThresholds, "fail", nil,
				"Set failure thresholds in the form `severity=count`.  The command will exit with exit code 2 if the assessment has count or more failed findings of the specified severity.")
		},
	}
}

func (o *AssessmentOpts) Validate() error {
	if err := o.ToolOpts.Validate(); err != nil {
		return err
	}
	if o.UploadEnabled {
		if err := o.RequireAuthentication(); err != nil {
			return err
		}
	}
	if len(o.FailThresholds) > 0 && !o.UploadEnabled {
		return fmt.Errorf("using --fail requires --upload=true")
	}
	parsedFailThresholds, err := assessments.ParseFailThresholds(o.FailThresholds)
	if err != nil {
		return err
	}
	o.parsedFailThresholds = parsedFailThresholds
	return nil
}

func ExtractArchives(dir string, archives []string) error {
	// policies dir by convention is where all policies are stored
	policiesDir := filepath.Join(dir, policy.Policies)

	fs := afero.NewOsFs()
	// remove all policies below e.g. dir/policies
	if err := fs.RemoveAll(policiesDir); err != nil {
		return err
	}
	for _, a := range archives {
		a = filepath.Join(dir, a)
		f, err := fs.Open(a)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}

		unpack := archive.Unzip
		// unpack the zip into dir, policies dir by convention will be unzipped
		// recreating dir/policies
		err = unpack(f, afero.NewBasePathFs(fs, dir), nil)
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
