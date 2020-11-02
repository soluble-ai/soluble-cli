// +build !ci

package resources

import (
	"net/http"
	"os"

	"github.com/soluble-ai/soluble-cli/pkg/util"
)

var FileSystem http.FileSystem

func init() {
	var root string
	// normally when running locallly we're somewhere in the project
	// directory tree in which case we use git to figure out the root
	// dir.  But if we're running outside that then  we'll have to set
	// an env variable
	if d := os.Getenv("__cli_root_dir__"); d != "" {
		root = d
	} else {
		// go tests run in the directory the test is in, so we use
		// git to tell us where the root directory is
		var err error
		root, err = util.Git("rev-parse", "--show-toplevel")
		if err != nil {
			panic(err)
		}
	}
	name := root + "/resources"
	FileSystem = http.Dir(name)
	rootPath = name
}
