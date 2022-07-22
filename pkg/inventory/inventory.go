// Copyright 2021 Soluble Inc
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

package inventory

import (
	"os"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type Manifest struct {
	root                          string
	TerraformRootModules          util.StringSet `json:"terraform_root_modules"`
	TerraformModules              util.StringSet `json:"terraform_modules"`
	CloudformationFiles           util.StringSet `json:"cloudformation_files"`
	HelmCharts                    util.StringSet `json:"helm_charts"`
	KubernetesManifestDirectories util.StringSet `json:"kubernetes_manifest_directories"`
	KustomizeDirectories          util.StringSet `json:"kustomize_directories"`
	CISystems                     util.StringSet `json:"ci_systems"`
	DockerDirectories             util.StringSet `json:"docker_directories"`
	Dockerfiles                   util.StringSet `jsont:"dockerfiles"`
	GODirectories                 util.StringSet `json:"go_directories"`
	PythonDirectories             util.StringSet `json:"python_directories"`
	NodeDirectories               util.StringSet `json:"node_directories"`
	JavaDirectories               util.StringSet `json:"java_directories"`
	RubyDirectories               util.StringSet `json:"ruby_directories"`
	CDKDirectories                util.StringSet `json:"cdk_directories"`
}

type FileDetector interface {
	DetectFileName(m *Manifest, path string) ContentDetector
}

type ContentDetector interface {
	DetectContent(*Manifest, string, []byte)
}

type DirDetector interface {
	DetectDirName(m *Manifest, path string)
}

type FinalizeDetector interface {
	FinalizeDetection(m *Manifest)
}

var cache = util.NewCache(3)

func (m *Manifest) getDetectors(detectors []interface{}) (fds []FileDetector, dds []DirDetector) {
	for _, d := range detectors {
		if fd, ok := d.(FileDetector); ok {
			fds = append(fds, fd)
		}
		if dd, ok := d.(DirDetector); ok {
			dds = append(dds, dd)
		}
	}
	return
}

func (m *Manifest) scan(root string, detectors ...interface{}) {
	root, _ = filepath.Abs(root)
	fileDetectors, dirDetectors := m.getDetectors(detectors)
	_ = filepath.WalkDir(root, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			log.Warnf("Could not scan {info:%s}: {warning:%s}", path, err)
			return nil
		}
		if info.IsDir() && info.Name() == ".git" {
			// skip .git directory
			return filepath.SkipDir
		}
		if isdir := info.IsDir(); isdir || info.Type().IsRegular() {
			relpath := path
			if filepath.IsAbs(relpath) {
				relpath, _ = filepath.Rel(root, relpath)
			}
			var cds []ContentDetector
			if isdir {
				for _, dd := range dirDetectors {
					dd.DetectDirName(m, relpath)
				}
			} else {
				for _, fd := range fileDetectors {
					if cd := fd.DetectFileName(m, relpath); cd != nil {
						cds = append(cds, cd)
					}
				}
			}
			if len(cds) > 0 {
				buf, err := os.ReadFile(path)
				if err != nil {
					log.Warnf("Could not read {info:%s}: {warning:%s}", path, err)
					return nil
				}
				if len(buf) > 0 {
					for _, d := range cds {
						d.DetectContent(m, relpath, buf)
					}
				}
			}
		}
		return nil
	})
	for _, d := range detectors {
		if fd, ok := d.(FinalizeDetector); ok {
			fd.FinalizeDetection(m)
		}
	}
}

func Do(root string) *Manifest {
	return cache.Get(root, func(dir string) interface{} {
		m := &Manifest{
			root: root,
		}
		m.scan(root,
			cloudformationDetector(0),
			kubernetesDetector(0),
			cidetector(0),
			dockerDetector(0),
			&terraformDetector{},
			goDetector(),
			pythonDetector(),
			javaAntMavenDetector(),
			javaGradleDetector(),
			nodeDetector(),
			rubyDetector(),
			cdkDetector(),
		)
		return m
	}).(*Manifest)
}
