package terraform

import (
	"fmt"
	"path/filepath"

	hclConfigs "github.com/hashicorp/terraform/configs"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/provider/output"
	"github.com/spf13/afero"
)

var (
	errLoadConfigFile = fmt.Errorf("failed to load config file")
)

// LoadIacFile parses the given terraform file from the given file path
func (s *Terraform) LoadIacFile(absFilePath string) (allResourcesConfig output.AllResourceConfigs, err error) {
	// new terraform config parser
	parser := hclConfigs.NewParser(afero.NewOsFs())

	hclFile, diags := parser.LoadConfigFile(absFilePath)
	if diags != nil {
		log.Errorf("failed to load config file '%s'. error:\n%v\n", absFilePath, diags)
		return allResourcesConfig, errLoadConfigFile
	}
	if hclFile == nil && diags.HasErrors() {
		log.Errorf("error occurred while loading config file. error:\n%v\n", diags)
		return allResourcesConfig, errLoadConfigFile
	}

	// initialize normalized output
	allResourcesConfig = make(map[string][]output.ResourceConfig)

	// traverse through all current's resources
	for _, managedResource := range hclFile.ManagedResources {
		// create output.ResourceConfig from hclConfigs.Resource
		resourceConfig, err := CreateResourceConfig(managedResource)
		if err != nil {
			return allResourcesConfig, fmt.Errorf("failed to create ResourceConfig")
		}

		// extract file name from path
		resourceConfig.Source = getFileName(resourceConfig.Source)

		// append to normalized output
		if _, present := allResourcesConfig[resourceConfig.Type]; !present {
			allResourcesConfig[resourceConfig.Type] = []output.ResourceConfig{resourceConfig}
		} else {
			allResourcesConfig[resourceConfig.Type] = append(allResourcesConfig[resourceConfig.Type], resourceConfig)
		}
	}

	// successful
	return allResourcesConfig, nil
}

// getFileName return file name from the given file path
func getFileName(path string) string {
	_, file := filepath.Split(path)
	return file
}
