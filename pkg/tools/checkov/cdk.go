package checkov

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type CDK struct {
	tools.DirectoryBasedToolOpts
	OutDirectory string
	SynthArgs    []string
	Synth        bool
}

var _ tools.Interface = (*CDK)(nil)

func (cdk *CDK) Name() string {
	return "checkov-cdk"
}

func (cdk *CDK) Register(cmd *cobra.Command) {
	cdk.DirectoryBasedToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVar(&cdk.OutDirectory, "cdk-out", "cdk.out", "The CDK output directory.")
	flags.BoolVar(&cdk.Synth, "cdk-synth", true, "Run \"cdk synth\" before scanning. Use --cdk-synth=false to disable.")
	flags.StringSliceVar(&cdk.SynthArgs, "cdk-synth-args", nil, "Pass these arguments to \"cdk synth\".  Comma separated, may be repeated.")
}

func (cdk *CDK) Validate() error {
	if err := cdk.DirectoryBasedToolOpts.Validate(); err != nil {
		return err
	}
	if !util.FileExists(filepath.Join(cdk.GetDirectory(), "cdk.json")) {
		return fmt.Errorf("cannot run cdk-scan in %s because it's not a CDK directory", cdk.GetDirectory())
	}
	cdkOut := cdk.getOutDirectory()
	if !cdk.Synth {
		if !util.DirExists(cdkOut) || util.DirEmpty(cdkOut) {
			return fmt.Errorf("cdk output directory %s does not exist or is empty", cdkOut)
		}
	}
	rel, err := filepath.Rel(cdk.GetDirectory(), cdkOut)
	if err != nil || strings.HasPrefix(rel, "..") {
		// The cdk.out directory must be a sub-directory of the directory we're running
		// in because we're using docker.
		return fmt.Errorf("cdk output directory %s must be relative to %s", cdk.OutDirectory, cdk.GetDirectory())
	}
	return nil
}

func (cdk *CDK) getOutDirectory() string {
	if filepath.IsAbs(cdk.OutDirectory) {
		return cdk.OutDirectory
	}
	return filepath.Join(cdk.GetDirectory(), cdk.OutDirectory)
}

func (cdk *CDK) Run() (*tools.Result, error) {
	if cdk.Synth {
		args := append([]string{"synth"}, cdk.SynthArgs...)
		if len(cdk.SynthArgs) == 0 {
			args = append(args, "--quiet")
		}
		synth := exec.Command("cdk", args...)
		synth.Dir = cdk.GetDirectory()
		synth.Stderr = os.Stderr
		synth.Stdout = os.Stderr
		exec := cdk.ExecuteCommand(synth)
		if !exec.ExpectExitCode(0) {
			log.Errorf("{primary:cdk synth} failed.  Run cdk synth manually and use {primary:--cdk-synth=false}.")
			return exec.ToResult(cdk.GetDirectory()), nil
		}
	}
	checkov := &Tool{
		DirectoryBasedToolOpts: cdk.DirectoryBasedToolOpts,
		Framework:              "cloudformation",
	}
	checkov.Directory = cdk.getOutDirectory()
	if err := checkov.Validate(); err != nil {
		return nil, err
	}
	result, err := checkov.Run()
	if result != nil {
		// all cdk findings are in generated files
		for _, f := range result.Findings {
			f.GeneratedFile = true
		}
	}
	return result, err
}
