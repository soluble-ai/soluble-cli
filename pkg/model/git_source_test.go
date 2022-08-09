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

package model

import (
	"os"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/repotree"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestGitSource(t *testing.T) {
	root, err := repotree.FindRepoRoot(".")
	util.Must(err)
	saveDir := config.ConfigDir
	config.ConfigDir, err = os.MkdirTemp("", "test-git-source*")
	util.Must(err)
	defer func() { _ = os.RemoveAll(config.ConfigDir); config.ConfigDir = saveDir }()
	assert := assert.New(t)
	s, err := GetGitSource("file://" + root)
	assert.Nil(err)
	assert.NotNil(s)
	f, err := s.GetFileSystem().Open("cmd/root/models/aws.hcl")
	assert.NoError(err)
	assert.NotNil(f)
	f.Close()
}
