package util

import (
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	_, err := os.Stat(filepath.FromSlash(path))
	return err == nil
}
