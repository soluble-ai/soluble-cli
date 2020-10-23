package iacinventory

type IacInventorier interface {
	Run() ([]GithubRepo, error)
}

func New(inventoryType interface{}) IacInventorier {
	return inventoryType.(IacInventorier)
}
