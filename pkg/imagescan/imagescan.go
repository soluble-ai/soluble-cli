package imagescan

import (
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/client"
)

const (
	trivy = "trivy"
)

type ImageScanner interface {
	Name() string
	Run() (*Result, error)
}

type Config struct {
	ReportEnabled bool
	Organizaton   string
	APIClient     client.Interface
	Image         string
}

type Result struct {
	N            *jnode.Node
	PrintPath    []string
	PrintColumns []string
}

func New(config Config) ImageScanner {
	scanner := &StockTrivy{
		image: config.Image,
		name:  trivy,
	}
	return scanner
}
