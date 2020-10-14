// +build !ci

package resources

import (
	"net/http"

	"github.com/soluble-ai/soluble-cli/pkg/util"
)

var FileSystem http.FileSystem

func init() {
	// go tests run in the directory the test is in, so we use
	// git to tell us where the root directory is
	root, err := util.Git("rev-parse", "--show-toplevel")
	if err != nil {
		panic(err)
	}
	name := root + "/resources"
	FileSystem = http.Dir(name)
	rootPath = name
}
