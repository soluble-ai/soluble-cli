package iacinventory

type IacInventorier interface {
	Run() ([]Repo, error)
}

func New(inventoryType interface{}) IacInventorier {
	return inventoryType.(IacInventorier)
}
