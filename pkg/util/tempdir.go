package util

import (
	"os"
	"path/filepath"
)

// rootTempDir - A temp dir which can be used as the root dir for any tmp dirs or files needed
var rootTempDir string

// CreateRootTempDir - call only once from root command
func CreateRootTempDir() error {
	var err error
	if rootTempDir == "" {
		rootTempDir, err = os.MkdirTemp("", "tmp")
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveRootTempDir - remove the root tmp dir and all subdirs and files, should be called on exit
func RemoveRootTempDir() string {
	if rootTempDir != "" {
		_ = os.RemoveAll(rootTempDir)
		dir := rootTempDir
		rootTempDir = ""
		return dir
	}
	return ""
}

// GetRootTempDir - return the absolute path to the root temp dir
func GetRootTempDir() string {
	return rootTempDir
}

// GetTempFilePath - return a unique absolute path for a file with filename
func GetTempFilePath(filename string) (string, error) {
	// create a unique tmp dir
	dir, err := os.MkdirTemp(rootTempDir, "*")
	if err != nil {
		return "", err
	}
	// create the absolute file path
	path := filepath.Join(dir, filename)
	return path, nil
}
