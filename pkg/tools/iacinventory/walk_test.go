package iacinventory

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/archive"
	"github.com/stretchr/testify/assert"
)

const (
	cloudformationTarball = "./testdata/cloudformation.tar.gz"
	cloudformationCFFile  = "cf.json"
)

func TestIsCloudformation(t *testing.T) {
	assert := assert.New(t)
	dir, err := ioutil.TempDir("", "cftest*")
	assert.Nil(err)
	err = archive.Do(archive.Untar, cloudformationTarball, dir, &archive.Options{
		TruncateFileSize: 1 << 20,
		IgnoreSymLinks:   true,
	})
	assert.Nil(err)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			if info.Name() == cloudformationCFFile {
				assert.True(isCloudFormationFile(path, info))
			} else {
				assert.False(isCloudFormationFile(path, info))
			}
		}
		return nil
	})
	assert.Nil(err)
}

const (
	dockerfileTarball = "./testdata/docker.tar.gz"
	dockerfile        = "Dockerfile"
	dottedDockerfile  = "dot.Dockerfile"
)

func TestIsDockerFile(t *testing.T) {
	assert := assert.New(t)
	dir, err := ioutil.TempDir("", "dockertest*")
	assert.Nil(err)
	err = archive.Do(archive.Untar, dockerfileTarball, dir, &archive.Options{
		TruncateFileSize: 1 << 20,
		IgnoreSymLinks:   true,
	})
	assert.Nil(err)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			if info.Name() == dockerfile || info.Name() == dottedDockerfile {
				assert.True(isDockerFile(path, info))
			} else {
				assert.False(isDockerFile(path, info))
			}
		}
		return nil
	})
	assert.Nil(err)
}

const (
	terraformTarball = "./testdata/terraform.tar.gz"
	terraformFile    = "main.tf"
)

func TestIsTerraform(t *testing.T) {
	assert := assert.New(t)
	dir, err := ioutil.TempDir("", "terraformtest*")
	assert.Nil(err)
	err = archive.Do(archive.Untar, terraformTarball, dir, &archive.Options{
		TruncateFileSize: 1 << 20,
		IgnoreSymLinks:   true,
	})
	assert.Nil(err)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			if info.Name() == terraformFile {
				assert.True(isTerraformFile(path, info))
			} else {
				assert.False(isTerraformFile(path, info))
			}
		}
		return nil
	})
	assert.Nil(err)
}

const (
	k8sTarball      = "./testdata/k8s.tar.gz"
	k8sManifestFile = "namespace.yaml"
)

func TestIsK8sManifest(t *testing.T) {
	assert := assert.New(t)
	dir, err := ioutil.TempDir("", "k8smanifesttest*")
	assert.Nil(err)
	err = archive.Do(archive.Untar, k8sTarball, dir, &archive.Options{
		TruncateFileSize: 1 << 20,
		IgnoreSymLinks:   true,
	})
	assert.Nil(err)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			if info.Name() == k8sManifestFile {
				assert.True(isKubernetesManifest(path, info))
			} else {
				assert.False(isKubernetesManifest(path, info))
			}
		}
		return nil
	})
	assert.Nil(err)
}
