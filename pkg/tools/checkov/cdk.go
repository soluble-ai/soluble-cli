package checkov

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

var _ tools.Interface = (*Helm)(nil)

func (cdk *CDK) Name() string {
	return "checkov-cdk"
}

func (cdk *CDK) Register(cmd *cobra.Command) {
	cdk.DirectoryBasedToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVar(&cdk.OutDirectory, "cdk-out", "cdk.out", "The CDK output directory.")
	flags.BoolVar(&cdk.Synth, "cdk-synth", false, "Run \"cdk synth\" before scanning. ")
	flags.StringSliceVar(&cdk.SynthArgs, "cdk-synth-args", nil, "Pass these arguments to \"cdk synth\".  Command separated, may be repeated.")
}

func (cdk *CDK) Validate() error {
	if err := cdk.DirectoryBasedToolOpts.Validate(); err != nil {
		return err
	}
	if !util.FileExists(filepath.Join(cdk.GetDirectory(), "cdk.json")) {
		return fmt.Errorf("cannot run cdk-scan in %s because it's not a CDK directory", cdk.GetDirectory())
	}
	return nil
}

func (cdk *CDK) Run() (*tools.Result, error) {
	if cdk.Synth {
		args := append([]string{"synth"}, cdk.SynthArgs...)
		synth := exec.Command("cdk", args...)
		synth.Dir = cdk.GetDirectory()
		synth.Stderr = os.Stderr
		synth.Stdout = os.Stderr
		cdk.LogCommand(synth)
		if err := synth.Run(); err != nil {
			return nil, err
		}
	}
	checkov := &Tool{
		DirectoryBasedToolOpts: cdk.DirectoryBasedToolOpts,
		Framework:              "cloudformation",
	}
	checkov.Directory = filepath.Join(cdk.GetDirectory(), cdk.OutDirectory)
	if err := checkov.Validate(); err != nil {
		return nil, err
	}
	return checkov.Run()
}
