package iacinventory

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// walkCloudFormationDirs implements WalkFunc to search for CloudFormation files in a repository.
func walkCloudFormationDirs(path string, info os.FileInfo, err error) (string, error) {
	if err != nil {
		return "", err
	}
	// if it is anything other than a regular file, return
	if !info.Mode().IsRegular() {
		return "", nil
	}
	// Cloudformation files do not have a unique extension, and are *typically*
	// ".yaml" or ".json" by convention. However, sometimes organizations use
	// Jinja, Go, or some other utility to template their Cloudformation.
	//
	// If the file has a possible extension and contains the string 'AWSTemplateFormatVersion',
	// then it is *most likely* a CF file.

	const maxSizeForCloudFormationConsideration int64 = 5 << (10 * 2) // 5MB, which is VERY large for json/yaml data
	if info.Size() > maxSizeForCloudFormationConsideration {
		// Exit early - file disqualified due to size.
		return "", nil
	}

	if strings.HasSuffix(info.Name(), ".json") ||
		strings.HasSuffix(info.Name(), ".yaml") ||
		strings.HasSuffix(info.Name(), ".yml") ||
		strings.HasSuffix(info.Name(), ".template") {
		f, err := os.Open(filepath.Clean(path))
		if err != nil {
			return "", fmt.Errorf("error opening file during CloudFormation analysis: %w", err)
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if bytes.Contains(scanner.Bytes(), []byte("AWSTemplateFormatVersion")) {
				return path, nil
			}
		}
	}
	return "", nil
}

// walkTerraformDirs implements WalkFunc to search for directories that contain Terraform files.
func walkTerraformDirs(path string, info os.FileInfo, err error) (string, error) {
	if err != nil {
		return "", err
	}
	// if it is anything other than a regular file, return
	if !info.Mode().IsRegular() {
		return "", nil
	}

	// If the file ends with TF, the parent directory contains terraform files
	if strings.HasSuffix(info.Name(), ".tf") {
		return path, nil
	}
	return "", nil
}
