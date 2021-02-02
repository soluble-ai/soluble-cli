package inventory

import (
	"os"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/util"
)

var gitDirCache = util.NewCache(5)

func FindGitDir(dir string) (string, error) {
	v := gitDirCache.Get(dir, func(s string) interface{} {
		dir, err := findGitDir(s)
		if err != nil {
			return err
		}
		return dir
	})
	if e, ok := v.(error); ok {
		return "", e
	}
	return v.(string), nil
}

func findGitDir(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		gd := filepath.Join(dir, ".git")
		if info, err := os.Stat(gd); err == nil && info.IsDir() {
			cf := filepath.Join(gd, "config")
			if info, err := os.Stat(cf); err == nil && info.Mode().IsRegular() {
				return filepath.Abs(gd)
			}
		}
		dir = filepath.Join(dir, "..")
		if dir[len(dir)-1] == os.PathSeparator {
			return "", nil
		}
	}
}

func FindRepoRoot(dir string) (string, error) {
	dir, err := FindGitDir(dir)
	if err == nil {
		dir = filepath.Dir(dir)
	}
	return dir, err
}
