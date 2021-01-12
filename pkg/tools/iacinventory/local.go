package iacinventory

import (
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

type Local struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Interface = &Local{}

func (t *Local) Name() string {
	return "local-inventory"
}

func (t *Local) Register(cmd *cobra.Command) {
	t.Internal = true
	t.DirectoryBasedToolOpts.Register(cmd)
}

func (t *Local) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "local",
		Short: "Inventory a local directory for infrastructure-as-code",
	}
}

func (t *Local) Run() (*tools.Result, error) {
	log.Infof("Finding local infrastructure-as-code inventory under {primary:%s}", t.GetDirectory())
	m := t.GetInventory()
	n, _ := print.ToResult(m)
	r := &tools.Result{
		Data: n,
	}
	return r, nil
}
