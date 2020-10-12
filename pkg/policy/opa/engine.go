// Copyright 2020 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opa

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/provider/output"
	"github.com/soluble-ai/soluble-cli/pkg/results"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

var (
	errInitFailed = fmt.Errorf("failed to initialize OPA policy engine")
)

// NewEngine returns a new OPA policy engine
func NewEngine(policyPath string) (*Engine, error) {
	if len(policyPath) == 0 {
		policyPath = "../../policies/opa/rego/aws"
	}

	fmt.Println("Policy path")
	fmt.Println(policyPath)

	// opa engine struct
	engine := &Engine{}
	// initialize the engine
	if err := engine.Init(policyPath); err != nil {
		log.Errorf("failed to initialize OPA policy engine: %s", err.Error())
		return engine, errInitFailed
	}
	// successful
	return engine, nil
}

// LoadRegoMetadata Loads rego metadata from a given file
func (e *Engine) LoadRegoMetadata(metaFilename string) (*RegoMetadata, error) {
	// Load metadata file if it exists
	metadata, err := ioutil.ReadFile(metaFilename)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Errorf("failed to load rego metadata: %s", err.Error())
		}
		return nil, err
	}

	// Read metadata into struct
	regoMetadata := RegoMetadata{}
	if err = json.Unmarshal(metadata, &regoMetadata); err != nil {
		log.Errorf("failed to unmarshal rego metadata: %s", err.Error())
		return nil, err
	}
	return &regoMetadata, err
}

// loadRawRegoFilesIntoMap imports raw rego files into a map
func (e *Engine) loadRawRegoFilesIntoMap(currentDir string, regoDataList []*RegoData, regoFileMap *map[string][]byte) error {
	for i := range regoDataList {
		regoPath := filepath.Join(currentDir, regoDataList[i].Metadata.File)
		rawRegoData, err := ioutil.ReadFile(regoPath)
		if err != nil {
			log.Errorf("failed to load rego file: %s", err.Error())
			continue
		}

		// Load the raw rego into the map
		_, ok := (*regoFileMap)[regoPath]
		if ok {
			// Already loaded this file, so continue
			log.Debugf("skipping already loaded rego file: %s", regoPath)
			continue
		}

		(*regoFileMap)[regoPath] = rawRegoData
	}
	return nil
}

// LoadRegoFiles Loads all related rego files from the given policy path into memory
func (e *Engine) LoadRegoFiles(policyPath string) error {
	// Walk the file path and find all directories
	dirList, err := util.FindAllDirectories(policyPath)
	if err != nil {
		return err
	}

	if len(dirList) == 0 {
		return fmt.Errorf("no directories found for path %s", policyPath)
	}

	e.regoFileMap = make(map[string][]byte)
	e.regoDataMap = make(map[string]*RegoData)

	// Load rego data files from each dir
	// First, we read the metadata file, which contains info about the associated rego rule. The .rego file data is
	// stored in a map in its raw format.
	sort.Strings(dirList)
	for i := range dirList {
		// Find all files in the current dir
		var fileInfo []os.FileInfo
		fileInfo, err = ioutil.ReadDir(dirList[i])
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Debugf("error while searching for files: %s", dirList[i])
			}
			continue
		}

		// Load the rego metadata first (*.json)
		metadataFiles := util.FilterFileInfoBySuffix(&fileInfo, []string{RegoMetadataFileSuffix})
		if len(metadataFiles) == 0 {
			log.Debugf("no metadata files were found: %s", dirList[i])
			continue
		}

		var regoDataList []*RegoData
		for j := range metadataFiles {
			filePath := filepath.Join(dirList[i], *metadataFiles[j])
			var regoMetadata *RegoMetadata
			regoMetadata, err = e.LoadRegoMetadata(filePath)
			if err != nil {
				log.Errorf("error loading rego metadata: %s", err.Error())
				continue
			}

			// Perform some sanity checks
			if strings.Contains(regoMetadata.Name, ".") {
				log.Errorf("error loading rego metadata: rule name must not contain a dot character: %s", regoMetadata.Name)
				continue
			}

			// Check for default template variable values specified
			if _, ok := regoMetadata.TemplateArgs["name"]; !ok {
				if regoMetadata.TemplateArgs == nil {
					regoMetadata.TemplateArgs = make(map[string]interface{})
				}
				// Add reserved template variable values
				regoMetadata.TemplateArgs["name"] = regoMetadata.Name
			}

			regoData := RegoData{
				Metadata: *regoMetadata,
			}

			regoDataList = append(regoDataList, &regoData)
			e.stats.metadataFileCount++
		}

		// Read in raw rego data from associated rego files
		if err = e.loadRawRegoFilesIntoMap(dirList[i], regoDataList, &e.regoFileMap); err != nil {
			log.Errorf("error loading raw rego data with path %s: %s", dirList[i], err.Error())
			continue
		}

		for j := range regoDataList {
			e.stats.metadataCount++

			// Check if the template file exists
			templateFile := filepath.Join(dirList[i], regoDataList[j].Metadata.File)

			// Apply templates if available
			var templateData bytes.Buffer
			t := template.New("opa")
			_, err = t.Parse(string(e.regoFileMap[templateFile]))
			if err != nil {
				log.Errorf("unable to parse template %s with error: %s", regoDataList[j].Metadata.File, err.Error())
				continue
			}
			if err = t.Execute(&templateData, regoDataList[j].Metadata.TemplateArgs); err != nil {
				log.Errorf("unable to execute template %s with error: %s", regoDataList[j].Metadata.File, err.Error())
				continue
			}

			regoDataList[j].RawRego = templateData.Bytes()
			if regoDataList[j].RawRego == nil {
				log.Debugf("raw rego data was null after applying template %s", regoDataList[j].Metadata.File)
				continue
			}
			e.regoDataMap[regoDataList[j].Metadata.ReferenceID] = regoDataList[j]
		}
	}

	e.stats.ruleCount = len(e.regoDataMap)
	e.stats.regoFileCount = len(e.regoFileMap)
	log.Infof("loaded %d Rego rules from %d rego files (%d metadata files).", e.stats.ruleCount, e.stats.regoFileCount, e.stats.metadataFileCount)

	return err
}

// CompileRegoFiles Compiles rego files for faster evaluation
func (e *Engine) CompileRegoFiles() error {
	for k := range e.regoDataMap {
		compiler, err := ast.CompileModules(map[string]string{
			e.regoDataMap[k].Metadata.Name: string(e.regoDataMap[k].RawRego),
		})
		if err != nil {
			log.Errorf("error compiling rego files %s with error %s", e.regoDataMap[k].Metadata.Name, err.Error())
			return err
		}

		r := rego.New(
			rego.Query(RuleQueryBase+"."+e.regoDataMap[k].Metadata.Name),
			rego.Compiler(compiler),
		)

		// Create a prepared query that can be evaluated.
		query, err := r.PrepareForEval(e.context)
		if err != nil {
			log.Errorf("error creating prepared query &s with error %s", e.regoDataMap[k].Metadata.Name, err.Error())
			return err
		}
		e.regoDataMap[k].PreparedQuery = &query
	}

	return nil
}

// Init initializes the Opa engine
// Handles loading all rules, filtering, compiling, and preparing for evaluation
func (e *Engine) Init(policyPath string) error {
	e.context = context.Background()
	if err := e.LoadRegoFiles(policyPath); err != nil {
		log.Errorf("error loading rego files from %s with error %s", policyPath, err.Error())
		return err
	}

	err := e.CompileRegoFiles()
	if err != nil {
		log.Errorf("error compiling rego files from %s with error %s", policyPath, err.Error())
		return err
	}

	// initialize ViolationStore
	e.results.ViolationStore = results.NewViolationStore()

	return nil
}

// GetResults Fetches results from OPA engine policy evaluation
func (e *Engine) GetResults() policy.EngineOutput {
	return e.results
}

// reportViolation Add a violation for a given resource
func (e *Engine) reportViolation(regoData *RegoData, resource *output.ResourceConfig) {
	violation := results.Violation{
		RuleName:     regoData.Metadata.Name,
		Description:  regoData.Metadata.Description,
		RuleID:       regoData.Metadata.ReferenceID,
		Severity:     regoData.Metadata.Severity,
		Category:     regoData.Metadata.Category,
		RuleFile:     regoData.Metadata.File,
		RuleData:     regoData.RawRego,
		ResourceName: resource.Name,
		ResourceType: resource.Type,
		ResourceData: resource.Config,
		File:         resource.Source,
		LineNumber:   resource.Line,
	}

	switch severity := regoData.Metadata.Severity; strings.ToLower(severity) {
	case "high":
		e.results.ViolationStore.Count.HighCount++
	case "medium":
		e.results.ViolationStore.Count.MediumCount++
	case "low":
		e.results.ViolationStore.Count.LowCount++
	default:
		log.Warnf("invalid severity found in rule definition %s with severity %s", violation.RuleID, severity)
	}
	e.results.ViolationStore.Count.TotalCount++
	e.results.ViolationStore.AddResult(&violation)
}

// Evaluate Executes compiled OPA queries against the input JSON data
func (e *Engine) Evaluate(engineInput policy.EngineInput) (policy.EngineOutput, error) {
	// Keep track of how long it takes to evaluate the policies
	start := time.Now()

	// Evaluate the policy against each resource type
	for k := range e.regoDataMap {
		// Execute the prepared query.
		rs, err := e.regoDataMap[k].PreparedQuery.Eval(e.context, rego.EvalInput(engineInput.InputData))
		if err != nil {
			log.Warnf("failed to run prepared query %s", err.Error())
			continue
		}

		if len(rs) == 0 || len(rs[0].Expressions) == 0 {
			log.Debugf("query executed but found no matches")
			continue
		}

		resourceViolations := rs[0].Expressions[0].Value.([]interface{})
		if len(resourceViolations) == 0 {
			log.Debugf("query executed but found no violations")
			continue
		}
		// Report a violation for each resource returned by the policy evaluation
		for i := range resourceViolations {
			var resourceID string

			// The return values come in two categories--either a map[string]interface{} type, where the "Id" key
			// contains the resource ID, or a string type which is the resource ID. This resource ID is where a
			// violation was found
			switch res := resourceViolations[i].(type) {
			case map[string]interface{}:
				_, ok := res["Id"]
				if !ok {
					log.Warnf("no Id key found in resource map: %s", res)
					continue
				}

				_, ok = res["Id"].(string)
				if !ok {
					log.Warnf("id key was invalid: %s", res)
					continue
				}
				resourceID = res["Id"].(string)
			case string:
				resourceID = res
			default:
				log.Warnf("resource ID format was invalid: %s", res)
				continue
			}

			// Locate the resource details within the input map
			var resource *output.ResourceConfig
			resource, err = util.FindResourceByID(resourceID, engineInput.InputData)
			if err != nil {
				log.Errorf(err.Error())
				continue
			}
			if resource == nil {
				log.Warnf("resource was not found: %s", resourceID)
				continue
			}

			log.Debugf("violation found for rule with rego %s", string("\n")+string(e.regoDataMap[k].RawRego)+string("\n"))

			// Report the violation
			e.reportViolation(e.regoDataMap[k], resource)
		}
	}

	e.stats.runTime = time.Since(start)
	return e.results, nil
}
