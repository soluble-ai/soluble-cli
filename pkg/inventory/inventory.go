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

func (m *Manifest) scan(root string, detectors ...interface{}) {
	buf := make([]byte, 4096)
	root, _ = filepath.Abs(root)
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
			var ds []ContentDetector
			for _, d := range detectors {
				if isdir {
					if dd, ok := d.(DirDetector); ok {
						dd.DetectDirName(m, relpath)
					}
				} else if fd, ok := d.(FileDetector); ok {
					if cd := fd.DetectFileName(m, relpath); cd != nil {
						ds = append(ds, cd)
					}
				}
			}
			if !isdir {
				if len(ds) > 0 {
					// read the first 4k of the file
					n, err := readFileStart(path, buf)
					if err != nil && !errors.Is(err, io.EOF) {
						log.Warnf("Could not read {info:%s}: {warning:%s}", path, err)
						return nil
					}
					if n > 0 {
						for _, d := range ds {
							d.DetectContent(m, relpath, buf[0:n])
						}
					}
				}
			}
		}
		return nil
	})
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
	m := &Manifest{}
	m.scan(root, cloudformationDetector(0), kubernetesDetector(0), cidetector(0),
		dockerDetector(0), &terraformDetector{})
	return m
}
