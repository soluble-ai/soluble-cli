package inventory

import (
	"os"
	"path/filepath"
)

func FindGitDir() (string, error) {
	dir, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	for {
		gd := filepath.Join(dir, ".git")
		if info, err := os.Stat(gd); err == nil && info.IsDir() {
			cf := filepath.Join(gd, "config")
			if info, err := os.Stat(cf); err == nil && info.Mode().IsRegular() {
				return gd, nil
			}
		}
		dir = filepath.Clean(filepath.Join(dir, ".."))
		if dir[len(dir)-1] == os.PathSeparator {
			return "", nil
		}
	}
}

func FindRepoRoot() (dir string, err error) {
	dir, err = FindGitDir()
	if err == nil {
		dir = filepath.Dir(dir)
	}
	return
}
