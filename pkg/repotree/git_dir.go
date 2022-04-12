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

package repotree

import (
	"os"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/util"
)

var gitDirCache = util.NewCache(5)

func FindGitDir(dir string) (string, error) {
	v := gitDirCache.Get(dir, func(s string) interface{} {
		dir, err := findGitDir(s)
		if err != nil {
			return err
		}
		return dir
	})
	if e, ok := v.(error); ok {
		return "", e
	}
	return v.(string), nil
}

func findGitDir(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		gd := filepath.Join(dir, ".git")
		if info, err := os.Stat(gd); err == nil && info.IsDir() {
			cf := filepath.Join(gd, "config")
			if info, err := os.Stat(cf); err == nil && info.Mode().IsRegular() {
				return filepath.Abs(gd)
			}
		}
		dir = filepath.Join(dir, "..")
		if dir[len(dir)-1] == os.PathSeparator {
			return "", nil
		}
	}
}

func FindRepoRoot(dir string) (string, error) {
	dir, err := FindGitDir(dir)
	if err == nil && dir != "" {
		dir = filepath.Dir(dir)
	}
	return dir, err
}
