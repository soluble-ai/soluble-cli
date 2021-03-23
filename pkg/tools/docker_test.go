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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocker(t *testing.T) {
	if hasDocker() == nil {
		assert := assert.New(t)
		dt := &DockerTool{
			Image: "hello-world",
		}
		d, err := dt.run(true)
		assert.Nil(err)
		assert.Contains(string(d), "Hello from Docker!")
	}
}
