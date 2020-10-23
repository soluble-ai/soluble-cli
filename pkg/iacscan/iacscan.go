package iacscan

import (
	"fmt"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/client"
)

type IacScanner interface {
	Run() (*Result, error)
}

type Result struct {
	N            *jnode.Node
	PrintPath    []string
	PrintColumns []string
}

type Config struct {
	ReportEnabled bool
	Organizaton   string
	APIClient     client.Interface
	Directory     string
	ScannerType   string
}

func New(config Config) (IacScanner, error) {
	var scanner IacScanner
	switch config.ScannerType {
	case "terrascan":
		scanner = &StockTerrascan{
			directory: config.Directory,
			apiClient: config.APIClient,
		}
	default:
		return nil, fmt.Errorf("unknown scanner %s", config.ScannerType)
	}
	if config.ReportEnabled {
		scanner = &Reporter{
			Config:  config,
			scanner: scanner,
			module:  config.ScannerType,
		}
	}
	return scanner, nil
}
