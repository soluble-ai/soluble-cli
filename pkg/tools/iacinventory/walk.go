package iacinventory

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

func isCloudFormationFile(path string, info os.FileInfo) bool {
	// Cloudformation files do not have a unique extension, and are *typically*
	// ".yaml" or ".json" by convention. However, sometimes organizations use
	// Jinja, Go, or some other utility to template their Cloudformation.
	//
	// If the file has a possible extension and contains the string 'AWSTemplateFormatVersion',
	// then it is *most likely* a CF file.
	const maxSizeForCloudFormationConsideration int64 = 5 << (10 * 2) // 5MB, which is VERY large for json/yaml data
	if info.Size() > maxSizeForCloudFormationConsideration {
		// Exit early - file disqualified due to size.
		return false
	}

	if strings.HasSuffix(info.Name(), ".json") ||
		strings.HasSuffix(info.Name(), ".yaml") ||
		strings.HasSuffix(info.Name(), ".yml") ||
		strings.HasSuffix(info.Name(), ".template") {
		f, err := os.Open(filepath.Clean(info.Name()))
		if err != nil {
			return false
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if bytes.Contains(scanner.Bytes(), []byte("AWSTemplateFormatVersion")) {
				return true
			}
		}
	}
	return false
}

// isTerraformFile implements WalkFunc to search for directories that contain Terraform files.
func isTerraformFile(path string, info os.FileInfo) bool {
	return strings.HasSuffix(info.Name(), ".tf")
}
