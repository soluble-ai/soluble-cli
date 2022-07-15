package checkov

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

type Plan struct {
	tools.DirectoryBasedToolOpts
	Plan     string
	Atlantis bool
}

var _ tools.Single = (*Plan)(nil)

func (p *Plan) Name() string {
	return "checkov-terraform-plan"
}

func (p *Plan) Register(cmd *cobra.Command) {
	p.DirectoryBasedToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVar(&p.Plan, "plan", "", "Scan the JSON format plan in `file`")
	_ = cmd.MarkFlagRequired("plan")
	flags.BoolVar(&p.Atlantis, "atlantis", true, "Print the results in the markdown format required for Atlantis output")

	if p.Atlantis {
		template_path := filepath.FromSlash("templates/atlantis.txt")
		log.Infof("Template directory: %s", template_path)
		//p.Tool.GetToolOptions().PrintOpts.Template = "hello {{ len .}}"
	}
}

func (p *Plan) Validate() error {
	if p.Directory != "" {
		log.Warnf("The value of --directory will be ignored")
	}
	p.Directory = filepath.Dir(p.Plan)
	if err := p.DirectoryBasedToolOpts.Validate(); err != nil {
		return err
	}
	// The plan has to be in JSON format and not native terraform format.
	// Detect the latter and suggest how to get the JSON format plan.
	ok, err := isTerraformNativePlan(p.Plan)
	if err != nil {
		return err
	}
	if ok {
		log.Warnf("This plan file is in terraform format.  To convert to the required JSON format, run:")
		log.Warnf("  {primary:terraform show -json %s > %s.json}", p.Plan, p.Plan)
		log.Warnf("And re-run the scan with {primary:--plan %s.json}", p.Plan)
		return fmt.Errorf("plan file is not in JSON format")
	}
	return nil
}

func (p *Plan) Run() (*tools.Result, error) {
	checkov := &Tool{
		DirectoryBasedToolOpts: p.DirectoryBasedToolOpts,
		Framework:              "terraform_plan",
		targetFile:             p.Plan,
	}
	if err := checkov.Validate(); err != nil {
		return nil, err
	}
	return checkov.Run()
}

func isTerraformNativePlan(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	buf := make([]byte, 4)
	_, _ = f.Read(buf)
	if buf[3] == 0x4 && buf[2] == 0x3 && buf[1] == 0x4b && buf[0] == 0x50 {
		return true, nil
	}
	return false, nil
}
