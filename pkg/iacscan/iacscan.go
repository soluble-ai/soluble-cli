package iacscan

import "github.com/soluble-ai/go-jnode"

type IacScanner interface {
	Run() (*jnode.Node, error)
}

func New(scannerType interface{}) IacScanner {
	return scannerType.(IacScanner)
}
