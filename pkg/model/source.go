// Copyright 2020 Soluble Inc
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
	"crypto/sha256"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/soluble-ai/soluble-cli/pkg/version"
)

type Source interface {
	GetFileSystem() http.FileSystem
	GetPath(file string) string
	GetVersion(file string, content []byte) string
	String() string
	IsEmbedded() bool
}

type FileSystemSource struct {
	Filesystem http.FileSystem
	RootPath   string
	Embedded   bool
}

func (s *FileSystemSource) GetPath(name string) string {
	return s.RootPath + "/" + name
}

func (s *FileSystemSource) GetFileSystem() http.FileSystem {
	return s.Filesystem
}

func (s *FileSystemSource) GetVersion(name string, content []byte) string {
	return version.Version
}

func (s *FileSystemSource) IsEmbedded() bool {
	return s.Embedded
}

func (s *FileSystemSource) String() string {
	return s.RootPath
}

func GetModelDir(url string) (string, error) {
	hash := sha256.Sum256([]byte(url))
	name := fmt.Sprintf("%012x", hash[0:6])
	m, err := homedir.Expand("~/.soluble_cli_models")
	if err != nil {
		return "", err
	}
	return filepath.Join(m, name), nil
}
