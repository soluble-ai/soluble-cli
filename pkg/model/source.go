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
)

type Source interface {
	GetFileSystem() http.FileSystem
	GetPath(file string) string
	GetVersion(file string, content []byte) string
	String() string
}

type fileSystemSource struct {
	filesystem http.FileSystem
	rootPath   string
}

var (
	embeddedSource Source
)

func GetEmbeddedSource() Source {
	if embeddedSource == nil {
		embeddedSource = &fileSystemSource{
			filesystem: embeddedFS,
			rootPath:   "<internal>",
		}
	}
	return embeddedSource
}

func (s *fileSystemSource) GetPath(name string) string {
	return s.rootPath + name
}

func (s *fileSystemSource) GetFileSystem() http.FileSystem {
	return s.filesystem
}

func (s *fileSystemSource) GetVersion(name string, content []byte) string {
	h := sha256.Sum256(content)
	return fmt.Sprintf("%012x", h[0:6])
}

func (s *fileSystemSource) String() string {
	return s.rootPath
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
