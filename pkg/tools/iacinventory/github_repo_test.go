package iacinventory

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanTarball(t *testing.T) {
	assert := assert.New(t)
	dir, err := ioutil.TempDir("", "soluble-cli-iacwalktests*")
	assert.Nil(err)
	defer os.RemoveAll(dir)
	g := &GithubRepo{FullName: "testdata/tesdata"}
	assert.Nil(g.scanTarball(dir, "testdata/testdata.tar.gz"))
	assert.ElementsMatch(g.CloudformationDirs, []string{
		"cloudformation/dir_with_cf",
	})
	assert.ElementsMatch(g.DockerfileDirs, []string{
		"docker/dir_with_docker/dot",
		"docker/dir_with_docker/plain",
		".", // should detect "dockerfile.rootdirtest" at repository root
	})
	assert.ElementsMatch(g.K8sManifestDirs, []string{
		"k8s/dir_with_manifests",
	})
	assert.ElementsMatch(g.TerraformDirs, []string{
		"terraform/dir_with_tf",
	})
	assert.ElementsMatch(g.CISystems, []CI{CIGithub})
}
