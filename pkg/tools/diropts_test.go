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

package tools

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectoryOpts(t *testing.T) {
	assert := assert.New(t)
	o := &DirectoryBasedToolOpts{
		Exclude: []string{"dir.go", "/secrets/testdata/"},
	}
	assert.NoError(o.Validate())
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

func TestNoRepo(t *testing.T) {
	assert := assert.New(t)
	dir, err := ioutil.TempDir("", "soluble-cli*")
	assert.NoError(err)
	defer os.RemoveAll(dir)
	o := &DirectoryBasedToolOpts{
		DirectoryOpt: DirectoryOpt{Directory: dir},
	}
	assert.Equal("", o.RepoRoot)
	assert.False(o.IsExcluded("foo.txt"))
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
