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

package print

import (
	"encoding/json"

	"github.com/soluble-ai/go-jnode"
)

// ToResult converts a value to a jnode.Node
func ToResult(value interface{}) (*jnode.Node, error) {
	dat, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return jnode.FromJSON(dat)
}
