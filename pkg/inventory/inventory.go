package inventory

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type Manifest struct {
	TerraformRootModuleDirectories util.StringSet `json:"terraform_root_modules"`
	CloudformationFiles            util.StringSet `json:"cloudformation_files"`
	HelmCharts                     util.StringSet `json:"helm_charts"`
	KubernetesManifestDirectories  util.StringSet `json:"kubernetes_manifest_directories"`
	CISystems                      util.StringSet `json:"ci_systems"`
	DockerDirectories              util.StringSet `json:"docker_directories"`
	GODirectories                  util.StringSet `json:"go_directories"`
	PythonDirectories              util.StringSet `json:"python_directories"`
	NodeDirectories                util.StringSet `json:"node_directories"`
	JavaDirectories                util.StringSet `json:"java_directories"`
	RubyDirectories                util.StringSet `json:"ruby_directories"`
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
	buf := make([]byte, 4096)
	root, _ = filepath.Abs(root)
	fileDetectors, dirDetectors := m.getDetectors(detectors)
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Warnf("Could not scan {info:%s}: {warning:%s}", path, err)
			return nil
		}
		if info.IsDir() && info.Name() == ".git" {
			// skip .git directory
			return filepath.SkipDir
		}
		if isdir := info.IsDir(); isdir || info.Mode().IsRegular() {
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
				// read the first 4k of the file
				n, err := readFileStart(path, buf)
				if err != nil && !errors.Is(err, io.EOF) {
					log.Warnf("Could not read {info:%s}: {warning:%s}", path, err)
					return nil
				}
				if n > 0 {
					for _, d := range cds {
						d.DetectContent(m, relpath, buf[0:n])
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

func readFileStart(path string, buf []byte) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Read(buf)
}

func Do(root string) *Manifest {
	return cache.Get(root, func(dir string) interface{} {
		m := &Manifest{}
		m.scan(root,
			cloudformationDetector(0),
			kubernetesDetector(0),
			cidetector(0),
			dockerDetector(0),
			&terraformDetector{},
			goDetector(),
			pythonDetector(),
			javaDetector(),
			nodeDetector(),
			rubyDetector(),
		)
		return m
	}).(*Manifest)
}
