package checkov

import (
	"fmt"

	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
)

type Kubernetes struct {
	Tool
}

var _ tools.Interface = (*Kubernetes)(nil)

func (k *Kubernetes) Validate() error {
	k.Framework = "kubernetes"
	// ignore kustomize template fragments
	m := k.GetInventory()
	for _, dir := range inventory.CollapseNestedDirs(m.KustomizeDirectories) {
		k.Exclude = append(k.Exclude, fmt.Sprintf("/%s/", dir))
	}
	return k.Tool.Validate()
}
