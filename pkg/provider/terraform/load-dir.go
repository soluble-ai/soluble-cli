package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	hclConfigs "github.com/hashicorp/terraform/configs"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/provider/output"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/afero"
)

var (
	errEmptyTFConfigDir = fmt.Errorf("directory has no terraform files")
	errLoadConfigDir    = fmt.Errorf("failed to load terraform allResourcesConfig dir")
	errBuildTFConfigDir = fmt.Errorf("failed to build terraform allResourcesConfig")
)

// LoadIacDir starts traversing from the given rootDir and traverses through
// all the descendant modules present to create an output list of all the
// resources present in rootDir and descendant modules
func (*Terraform) LoadIacDir(absRootDir string) (allResourcesConfig output.AllResourceConfigs, err error) {

	// create a new config parser
	parser := hclConfigs.NewParser(afero.NewOsFs())

	// check if the directory has any tf config files (.tf or .tf.json)
	if !parser.IsConfigDir(absRootDir) {
		log.Errorf("directory '%s' has no terraform config files", absRootDir)
		return allResourcesConfig, errEmptyTFConfigDir
	}

	// load root config directory
	rootMod, diags := parser.LoadConfigDir(absRootDir)
	if diags.HasErrors() {
		log.Errorf("failed to load terraform config dir '%s'. error:\n%+v\n", absRootDir, diags)
		return allResourcesConfig, errLoadConfigDir
	}

	// create a new remote module installer to install remote modules
	r := NewRemoteModuleInstaller()
	defer r.CleanUp()

	// using the BuildConfig and ModuleWalkerFunc to traverse through all
	// descendant modules from the root module and create a unified
	// configuration of type *configs.Config
	// Note: currently, only Local paths are supported for Module Sources
	versionI := 0
	unified, diags := hclConfigs.BuildConfig(rootMod, hclConfigs.ModuleWalkerFunc(
		func(req *hclConfigs.ModuleRequest) (*hclConfigs.Module, *version.Version, hcl.Diagnostics) {

			// figure out path sub module directory, if it's remote then download it locally
			var pathToModule string
			if isLocalSourceAddr(req.SourceAddr) {
				// determine the absolute path from root module to the sub module
				// using *configs.ModuleRequest.Path field
				pathArr := strings.Split(req.Path.String(), ".")
				pathArr = pathArr[:len(pathArr)-1]
				pathToModule = filepath.Join(absRootDir, filepath.Join(pathArr...), req.SourceAddr)
				log.Debugf("processing local module %q", req.SourceAddr)
			} else {
				// temp dir to download the remote repo
				tempDir := filepath.Join(os.TempDir(), util.GenRandomString(6))

				// Download remote module
				pathToModule, err = r.DownloadModule(req.SourceAddr, tempDir)
				if err != nil {
					log.Errorf("failed to download remote module %q. error: '%v'", req.SourceAddr, err)
				}
			}

			// load sub module directory
			subMod, diags := parser.LoadConfigDir(pathToModule)
			version, _ := version.NewVersion(fmt.Sprintf("1.0.%d", versionI))
			versionI++
			return subMod, version, diags
		},
	))
	if diags.HasErrors() {
		log.Errorf("failed to build unified config. errors:\n%+v\n", diags)
		return allResourcesConfig, errBuildTFConfigDir
	}

	/*
		The "unified" config created from BuildConfig in the previous step
		represents a tree structure with rootDir module being at its root and
		all the sub modules being its children, and these children can have
		more children and so on...

		Now, using BFS we traverse through all the submodules using the classic
		approach of using a queue data structure
	*/

	// queue of for BFS, add root module config to it
	configsQ := []*hclConfigs.Config{unified.Root}

	// initialize normalized output
	allResourcesConfig = make(map[string][]output.ResourceConfig)

	// using BFS traverse through all modules in the unified config tree
	log.Debugf("traversing through all modules in config tree")
	for len(configsQ) > 0 {

		// pop first element from the queue
		current := configsQ[0]
		configsQ = configsQ[1:]

		// traverse through all current's resources
		for _, managedResource := range current.Module.ManagedResources {

			// create output.ResourceConfig from hclConfigs.Resource
			resourceConfig, err := CreateResourceConfig(managedResource)
			if err != nil {
				return allResourcesConfig, fmt.Errorf("failed to create ResourceConfig")
			}

			resourceConfig.Source, err = filepath.Rel(absRootDir, resourceConfig.Source)
			if err != nil {
				return allResourcesConfig, fmt.Errorf("failed to get resource: %s", err)
			}

			// append to normalized output
			if _, present := allResourcesConfig[resourceConfig.Type]; !present {
				allResourcesConfig[resourceConfig.Type] = []output.ResourceConfig{resourceConfig}
			} else {
				allResourcesConfig[resourceConfig.Type] = append(allResourcesConfig[resourceConfig.Type], resourceConfig)
			}
		}

		// add all current's children to the queue
		for _, childModule := range current.Children {
			configsQ = append(configsQ, childModule)
		}
	}

	// successful
	return allResourcesConfig, nil
}
