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

// +build !ci

package root

import (
	"net/http"

	"github.com/soluble-ai/soluble-cli/pkg/model"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

var embeddedFS http.FileSystem

func init() {
	// go tests run in the directory the test is in, so we use
	// git to tell us where the root directory is
	root, err := util.Git("rev-parse", "--show-toplevel")
	if err != nil {
		panic(err)
	}
	name := root + "/models"
	embeddedFS = http.Dir(name)
	embeddedModelsSource = &model.FileSystemSource{
		Filesystem: embeddedFS,
		RootPath:   name,
		Embedded:   true,
	}
}
