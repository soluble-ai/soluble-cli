package iacinventory

import "github.com/soluble-ai/go-jnode"

type IacInventorier interface {
	Run() (*jnode.Node, error)
	Stop() error
}

func New(inventoryType interface{}) IacInventorier {
	return inventoryType.(IacInventorier)
}
