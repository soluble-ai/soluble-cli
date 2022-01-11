package util

import (
	"errors"
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	_, err := os.Stat(filepath.FromSlash(path))
	return err == nil
}

func DirExists(path string) bool {
	di, err := os.Stat(filepath.FromSlash(path))
	if err == nil && di.IsDir() {
		return true
	}
	return false
}

func DirEmpty(path string) bool {
	es, err := os.ReadDir(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return true
	}
	if err == nil && len(es) == 0 {
		return true
	}
	return false
}
