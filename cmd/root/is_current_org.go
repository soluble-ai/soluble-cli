package root

import (
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/model"
)

func isCurrentOrgFunction(n *jnode.Node) interface{} {
	org := n.Path("orgId").AsText()
	if org == config.Get().Organization {
		return "*"
	}
	return ""
}

func init() {
	model.RegisterColumnFunction("is_current_org", isCurrentOrgFunction)
}
