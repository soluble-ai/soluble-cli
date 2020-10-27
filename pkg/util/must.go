package util

import "github.com/soluble-ai/soluble-cli/pkg/log"

// Must panics if err is non-nil
func Must(err error) {
	if err != nil {
		log.Errorf("must {danger:%s}", err)
		panic(err)
	}
}
