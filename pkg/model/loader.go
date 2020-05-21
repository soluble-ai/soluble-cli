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

package model

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/version"
)

type modelLoader struct {
	models      []*Model
	diagnostics hcl.Diagnostics
	parser      *hclparse.Parser
}

func Load(source Source) error {
	log.Debugf("Loading models from {info:%s}", source)
	m := &modelLoader{
		parser: hclparse.NewParser(),
	}
	if err := m.loadModels(source, ""); err != nil {
		return err
	}
	wr := hcl.NewDiagnosticTextWriter(
		os.Stdout,        // writer to send messages to
		m.parser.Files(), // the parser's file cache, for source snippets
		78,               // wrapping width
		true,             // generate colored/highlighted output
	)
	_ = wr.WriteDiagnostics(m.diagnostics)
	if m.diagnostics.HasErrors() {
		return fmt.Errorf("some models have errors")
	}
	for _, model := range m.models {
		if model.MinCLIVersion != nil && !version.IsCompatible(*model.MinCLIVersion) {
			log.Warnf("The model in %s is not compatible with this version of the CLI (require %s)",
				model.FileName, *model.MinCLIVersion)
		}
		if err := model.validate(); err != nil {
			return err
		}
		model.Source = source
	}
	Models = append(Models, m.models...)
	return nil
}

func (m *modelLoader) loadModels(source Source, dirName string) error {
	dir, err := source.GetFileSystem().Open(dirName)
	if err != nil {
		return err
	}
	defer dir.Close()
	fileInfos, err := dir.Readdir(0)
	if err != nil {
		return err
	}
	for _, fileInfo := range fileInfos {
		var path string
		if dirName == "" {
			path = fileInfo.Name()
		} else {
			path = dirName + "/" + fileInfo.Name()
		}
		if fileInfo.IsDir() {
			err := m.loadModels(source, path)
			if err != nil {
				return err
			}
		} else if strings.HasSuffix(fileInfo.Name(), ".hcl") {
			err := m.loadModel(source, path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *modelLoader) loadModel(source Source, name string) error {
	f, err := source.GetFileSystem().Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	src, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	file, diag := m.parser.ParseHCL(src, source.GetPath(name))
	m.diagnostics = m.diagnostics.Extend(diag)
	if !diag.HasErrors() {
		model := &Model{
			FileName: source.GetPath(name),
			Version:  source.GetVersion(name, src),
		}
		diag = gohcl.DecodeBody(file.Body, nil, model)
		m.diagnostics = m.diagnostics.Extend(diag)
		if !diag.HasErrors() {
			m.models = append(m.models, model)
			log.Debugf("%s defines command %s", model.FileName, model.Command.Name)
		}
	}
	return nil
}
