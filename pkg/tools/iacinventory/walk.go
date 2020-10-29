package iacinventory

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// fileTooLarge returns true if a file is too large perform a full read.
func fileTooLarge(info os.FileInfo) bool {
	const maxSize int64 = 5 << (10 * 2) // 5MB, which is VERY large for json/yaml data
	return info.Size() > maxSize
}

func IsCloudFormationFile(path string, info os.FileInfo) bool {
	// Cloudformation files do not have a unique extension, and are *typically*
	// ".yaml" or ".json" by convention. However, sometimes organizations use
	// Jinja, Go, or some other utility to template their Cloudformation.
	//
	// If the file has a possible extension and contains the string 'AWSTemplateFormatVersion',
	// then it is *most likely* a CF file.

	if fileTooLarge(info) {
		return false
	}

	if !strings.HasSuffix(info.Name(), ".json") &&
		!strings.HasSuffix(info.Name(), ".yaml") &&
		!strings.HasSuffix(info.Name(), ".yml") &&
		!strings.HasSuffix(info.Name(), ".template") {
		return false
	}

	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	// If the resource contains `AWSTemplateFormatVersion`, it is almost certainly
	// a cloudformation file.
	// If the resource did not contain `AWSTemplateFormatVersion`, it may still be
	// a valid cloudformation file.
	var hasResources bool
	var hasAWScoloncolon bool
	for scanner.Scan() {
		if bytes.Contains(scanner.Bytes(), []byte("AWSTemplateFormatVersion")) {
			return true
		}
		if bytes.Contains(scanner.Bytes(), []byte("Resources")) {
			hasResources = true
		}
		if bytes.Contains(scanner.Bytes(), []byte("AWS::")) {
			hasAWScoloncolon = true
		}
	}
	if hasResources && hasAWScoloncolon {
		return true
	}
	return false
}

// IsTerraformFile implements WalkFunc to search for directories that contain Terraform files.
func IsTerraformFile(_ string, info os.FileInfo) bool {
	return strings.HasSuffix(info.Name(), ".tf")
}

// IsDockerFile returns true if a file is a dockerfile.
func IsDockerFile(path string, info os.FileInfo) bool {
	if fileTooLarge(info) {
		return false
	}
	// file should contain "dockerfile"
	if !strings.Contains(strings.ToUpper(info.Name()), "DOCKERFILE") {
		return false
	}
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// and should always contain a FROM directive.
		if bytes.Contains(scanner.Bytes(), []byte("FROM ")) {
			return true
		}
	}
	return false
}

func IsKubernetesManifest(path string, info os.FileInfo) bool {
	if fileTooLarge(info) {
		return false
	}

	// file must end in '.yaml', '.yml', or '.json'
	if !strings.HasSuffix(info.Name(), ".yaml") &&
		!strings.HasSuffix(info.Name(), ".yml") &&
		!strings.HasSuffix(info.Name(), ".json") {
		return false
	}
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)

	// special case: helm charts
	if info.Name() == "Chart.yaml" {
		for scanner.Scan() {
			if bytes.Contains(scanner.Bytes(), []byte("apiVersion")) {
				return true
			}
		}
	}

	// regular k8s manifests
	manifestRequiredFields := map[string]bool{
		"apiVersion": false,
		"kind":       false,
		"metadata":   false,
		"spec":       false,
	}
	for scanner.Scan() {
		// and should always contain a FROM directive.
		for k := range manifestRequiredFields {
			if bytes.Contains(scanner.Bytes(), []byte(k)) {
				manifestRequiredFields[k] = true
			}
		}
	}
	for _, v := range manifestRequiredFields {
		// if a required key was missing
		if !v {
			return false
		}
	}
	return true
}
