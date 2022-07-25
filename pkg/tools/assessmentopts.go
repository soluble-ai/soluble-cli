package tools

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/exit"
	"github.com/soluble-ai/soluble-cli/pkg/log"
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

	parsedFailThresholds map[string]int
	customPoliciesDir    *string
	customPolicyMetadata map[string]string
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
		Example: `
A tool run can optionally exit with exit code 2 if the assessment contains
failed findings.  For example:
		
# Fail if 1 or more high or critical severity findings in this build:
soluble ... --fail high=1
# Or shorter:
soluble ... --fail high

The severity levels are critical, high, medium, low, and info in that order.`,
		CreateFlagsFunc: func(flags *pflag.FlagSet) {
			flags.BoolVar(&o.DisableCustomPolicies, "disable-custom-policies", false, "Don't use custom policies")
			flags.StringVar(&o.CustomPoliciesDir, "custom-policies", "", "Use custom policies from `dir`.")
			flags.BoolVar(&o.PrintResultOpt, "print-result", false, "Print the JSON result from the tool on stderr")
			flags.StringVar(&o.SaveResult, "save-result", "", "Save the JSON reesult from the tool to `file`")
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
		if err := o.RequireAPIToken(); err != nil {
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

func (o *AssessmentOpts) GetCustomPoliciesDir() (string, error) {
	if o.PreparedCustomPoliciesDir != "" {
		return o.PreparedCustomPoliciesDir, nil
	}
	if o.DisableCustomPolicies {
		return "", nil
	}
	if o.customPoliciesDir != nil {
		return *o.customPoliciesDir, nil
	}
	if o.GetAPIClientConfig().APIToken == "" {
		return "", nil
	}
	dir := o.CustomPoliciesDir
	if dir == "" {
		url := fmt.Sprintf("/api/v1/org/{org}/custom/policies/%s/rules.tgz", o.Tool.Name())
		d, err := o.InstallAPIServerArtifact(fmt.Sprintf("%s-%s-policies", o.Tool.Name(),
			o.GetAPIClientConfig().Organization), url)
		if err != nil {
			return "", err
		}
		dir = d.Dir
	}
	// if the directory is empty, then treat that the same as no custom policies
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}
	if len(fs) == 0 {
		var zero string
		o.customPoliciesDir = &zero
		log.Infof("{primary:%s} has no custom policies", o.Tool.Name())
	} else {
		store := &policy.Store{Dir: dir}
		dest, err := os.MkdirTemp("", "policy*")
		if err != nil {
			return "", err
		}
		exit.AddFunc(func() { _ = os.RemoveAll(dest) })
		if err := store.LoadRules(); err != nil {
			return "", err
		}
		if err := store.PrepareRules(dest); err != nil {
			return "", err
		}
		md, err := store.GetPolicyUploadMetadata()
		if err != nil {
			return "", err
		}
		o.customPolicyMetadata = md
		o.customPoliciesDir = &dest
	}
	return *o.customPoliciesDir, nil
}
