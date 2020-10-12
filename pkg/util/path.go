package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/log"
)

// GetAbsPath returns absolute path from passed file path resolving even ~ to user home dir and any other such symbols that are only
// shell expanded can also be handled here
func GetAbsPath(path string) (string, error) {
	// Only shell resolves `~` to home so handle it specially
	if strings.HasPrefix(path, "~") {
		homeDir := os.Getenv("HOME")
		if len(path) > 1 {
			path = filepath.Join(homeDir, path[1:])
		} else {
			return homeDir, nil
		}
	}

	// get absolute file path
	path, _ = filepath.Abs(path)
	return path, nil
}

// FindAllDirectories Walks the file path and returns a list of all directories within
func FindAllDirectories(basePath string) ([]string, error) {
	dirList := make([]string, 0)
	err := filepath.Walk(basePath, func(filePath string, fileInfo os.FileInfo, err error) error {
		if fileInfo != nil && fileInfo.IsDir() {
			dirList = append(dirList, filePath)
		}
		return err
	})
	return dirList, err
}

// FilterFileInfoBySuffix Given a list of files, returns a subset of files containing a suffix which matches the input filter
func FilterFileInfoBySuffix(allFileList *[]os.FileInfo, filter []string) []*string {
	fileList := make([]*string, 0)

	for i := range *allFileList {
		for j := range filter {
			if strings.HasSuffix((*allFileList)[i].Name(), filter[j]) {
				filename := (*allFileList)[i].Name()
				fileList = append(fileList, &filename)
			}
		}
	}
	return fileList
}

// FindFilesBySuffix finds all files within a given directory that have the specified suffixes
// Returns a map with keys as directories and values as a list of files
func FindFilesBySuffix(basePath string, suffixes []string) (map[string][]*string, error) {
	retMap := make(map[string][]*string)

	// Walk the file path and find all directories
	dirList, err := FindAllDirectories(basePath)
	if err != nil {
		log.Errorf("error encountered traversing directories %s with error %s", basePath, err.Error())
		return retMap, err
	}

	if len(dirList) == 0 {
		return retMap, fmt.Errorf("no directories found for path %s", basePath)
	}

	sort.Strings(dirList)
	for i := range dirList {
		// Find all files in the current dir
		var fileInfo []os.FileInfo
		fileInfo, err = ioutil.ReadDir(dirList[i])
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Debugf("error while searching for files: %s failed with error: %s", dirList[i], err.Error())
			}
			continue
		}

		fileList := FilterFileInfoBySuffix(&fileInfo, suffixes)
		if len(fileList) > 0 {
			retMap[dirList[i]] = fileList
		}
	}

	return retMap, nil
}
