package iacinventory

import (
	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

type Local struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Interface = &Local{}

func (t *Local) Name() string {
	return "local-inventory"
}

func (t *Local) Run() (*tools.Result, error) {
	log.Infof("Finding local infrastructure-as-code inventory under {primary:%s}", t.GetDirectory())
	m := inventory.Do(t.GetDirectory())
	n, _ := print.ToResult(m)
	r := &tools.Result{
		Data: n,
	}
	return r, nil
}
