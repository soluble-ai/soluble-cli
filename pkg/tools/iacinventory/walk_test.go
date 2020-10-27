package iacinventory

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/archive"
	"github.com/stretchr/testify/assert"
)

var tests = []struct {
	tarball         string              // file to tarball path on disk (provided by go generate)
	expectedMatches map[string][]string // map[file][]iactypes
}{
	{
		"./testdata/cloudformation.tar.gz",
		map[string][]string{
			"./cloudformation/dir_with_cf/cf.json":    {"cloudformation"},
			"./cloudformation/dir_without_cf/nothere": {},
		},
	},
	{
		"./testdata/docker.tar.gz",
		map[string][]string{
			"./docker/dir_with_docker/Dockerfile":     {"docker"},
			"./docker/dir_with_docker/dot.Dockerfile": {"docker"},
			"./docker/dir_without_docker/nodocker":    {},
		},
	},
	{
		"./testdata/terraform.tar.gz",
		map[string][]string{
			"./terraform/dir_with_tf/main.tf":      {"terraform"},
			"./terraform/dir_without_tf/no_tf.txt": {},
		},
	},
	{
		"./testdata/k8s.tar.gz",
		map[string][]string{
			"./k8s/dir_with_manifests/namespace.yaml":     {"kubernetes"},
			"./k8s/dir_with_manifests/Chart.yaml":         {"kubernetes"}, // "helm"
			"./k8s/dir_without_manifests/incomplete.yaml": {},
			"./k8s/dir_without_manifests/not-a-manifest":  {},
		},
	},
}

func TestIACWalks(t *testing.T) {
	assert := assert.New(t)
	dir, err := ioutil.TempDir("", "soluble-cli-iacwalktests*")
	assert.Nil(err)
	for _, test := range tests {
		err = archive.Do(archive.Untar, test.tarball, dir, &archive.Options{
			TruncateFileSize: 1 << 20,
			IgnoreSymLinks:   true,
		})
		assert.Nil(err)
		matches := make(map[string][]string) // map[file][]iacType
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Mode().IsRegular() {
				f := "." + strings.TrimPrefix(path, dir)
				if isCloudFormationFile(path, info) {
					types := matches[info.Name()]
					types = append(types, "cloudformation")
					matches[f] = types
				}
				if isDockerFile(path, info) {
					types := matches[info.Name()]
					types = append(types, "docker")
					matches[f] = types
				}
				if isTerraformFile(path, info) {
					types := matches[info.Name()]
					types = append(types, "terraform")
					matches[f] = types
				}
				if isKubernetesManifest(path, info) {
					types := matches[info.Name()]
					types = append(types, "kubernetes")
					matches[f] = types
				}
			}
			return nil
		})
		assert.Nil(err)
		for file, expectedIACs := range test.expectedMatches {
			iacs, ok := matches[file]
			if len(expectedIACs) == 0 {
				if ok {
					t.Fatalf("expected no matches for for file %q in tarball %q, but got matches: %q", file, test.tarball, iacs)
				}
				continue
			}
			if !ok {
				t.Fatalf("expected a match of %q for file %q in tarball %q, but got no matches", expectedIACs, file, test.tarball)
			}
			assert.EqualValues(expectedIACs, iacs)
		}
	}
}
