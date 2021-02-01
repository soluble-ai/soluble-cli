package tools

import (
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/log"
)

func MustRel(base, target string) string {
	if !filepath.IsAbs(target) {
		var err error
		target, err = filepath.Abs(target)
		if err != nil {
			log.Errorf("Could not determine absolute path of {warning:%s} - {danger:s}", target, err)
			panic(err)
		}
	}
	rel, err := filepath.Rel(base, target)
	if err != nil {
		log.Errorf("Could not determine relative path of {warning:%s} and {warning:%s} - {danger:%s}", base, target, err)
		panic(err)
	}
	return rel
}
