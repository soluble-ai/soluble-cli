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

//go:build !ci
// +build !ci

package version

import (
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

func init() {
	// When running locally initialize version strings by looking at
	// local git version. But first make sure we're in the soluble-cli
	// repository.
	u, err := util.Git("config", "remote.origin.url")
	if err == nil {
		if u == "git@github.com:soluble-ai/soluble-cli.git" {
			v, err := util.Git("describe", "--tags", "--dirty", "--always")
			if err == nil {
				Version = v
			}
		}
	}
}
