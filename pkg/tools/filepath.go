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
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/log"
)

func MustRel(base, target string) string {
	if !filepath.IsAbs(target) {
		var err error
		target, err = filepath.Abs(target)
		if err != nil {
			log.Errorf("Could not determine absolute path of {warning:%s} - {danger:s}", target, err)
			panic(err)
		}
	}
	rel, err := filepath.Rel(base, target)
	if err != nil {
		log.Errorf("Could not determine relative path of {warning:%s} and {warning:%s} - {danger:%s}", base, target, err)
		panic(err)
	}
	return rel
}
