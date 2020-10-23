package iacscan

import (
	"bytes"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
)

type Reporter struct {
	Config
	scanner IacScanner
	module  string
}

func (r *Reporter) Run() (*Result, error) {
	result, err := r.scanner.Run()
	if err != nil {
		return nil, err
	}
	rr := bytes.NewReader([]byte(result.N.String()))
	log.Infof("Uploading iac scan results")
	values := map[string]string{
		"directory":   r.Directory,
		"scannerType": r.ScannerType,
	}
	err = r.APIClient.XCPPost(r.Organizaton, r.module, nil, values,
		xcp.WithCIEnv, xcp.WithReader("results_json", "results.json", rr))
	if err != nil {
		return nil, err
	}
	return result, nil
}
