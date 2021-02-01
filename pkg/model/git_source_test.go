package model

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestGitSource(t *testing.T) {
	root, err := inventory.FindRepoRoot(".")
	util.Must(err)
	saveDir := config.ConfigDir
	config.ConfigDir, err = ioutil.TempDir("", "test-git-source*")
	util.Must(err)
	defer func() { _ = os.RemoveAll(config.ConfigDir); config.ConfigDir = saveDir }()
	assert := assert.New(t)
	s, err := GetGitSource(root)
	assert.Nil(err)
	assert.NotNil(s)
	f, err := s.GetFileSystem().Open("resources/models/aws.hcl")
	assert.Nil(err)
	f.Close()
}
