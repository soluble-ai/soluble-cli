package tools

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectoryOpts(t *testing.T) {
	assert := assert.New(t)
	o := &DirectoryBasedToolOpts{
		Exclude: []string{"dir.go"},
	}
	assert.Nil(o.Validate())
	dir := o.GetDirectory()
	assert.True(filepath.IsAbs(dir))
	assert.NotEmpty(o.RepoRoot)
	assert.NotNil(o.GetConfig())
	assert.True(o.IsExcluded("secrets/testdata/results.json"))
	files, err := o.GetFilesInDirectory([]string{"dir_test.go"})
	assert.Nil(err)
	assert.Equal([]string{"dir_test.go"}, files)
	assert.Equal([]string{"dir_test.go"}, o.RemoveExcluded([]string{"dir.go", "dir_test.go"}))
}

func TestDirectoryOptsExclude(t *testing.T) {
	assert := assert.New(t)
	o := &DirectoryBasedToolOpts{
		Exclude: []string{"go.sum"},
	}
	assert.Nil(o.Validate())
	o.Directory = o.RepoRoot
	assert.Empty(o.RemoveExcluded([]string{filepath.Join(o.Directory, "go.sum")}))
}

func TestGetInventory(t *testing.T) {
	assert := assert.New(t)
	o := &DirectoryBasedToolOpts{}
	m := o.GetInventory()
	assert.NotNil(m)
}
