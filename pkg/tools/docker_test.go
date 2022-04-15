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
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestNoDocker(t *testing.T) {
	assert := assert.New(t)
	assert.False(IsDockerError(nil))
	assert.False(IsDockerError(fmt.Errorf("not a docker error")))
	f, err := util.TempFile("docker*")
	if assert.NoError(err) {
		defer os.Remove(f)
		err := hasDocker(func(c *exec.Cmd) {
			c.Args = []string{"docker", "-H", fmt.Sprintf("unix://%s", f), "info"}
		})
		assert.Error(err)
		assert.True(IsDockerError(err))
	}
}

func TestDocker(t *testing.T) {
	if hasDocker() == nil {
		assert := assert.New(t)
		dt := &DockerTool{
			Image: "hello-world",
		}
		res, err := dt.run(true)
		assert.NoError(err)
		assert.Contains(string(res.Output), "Hello from Docker!")
		assert.Contains(res.CombinedOutput, "Hello from Docker!")
	}
}

func TestDockerGetArgs(t *testing.T) {
	assert := assert.New(t)
	dt := &DockerTool{
		DockerArgs: []string{"-m", "256m"},
		Image:      "test",
		Args:       []string{"arg1"},
		Directory:  "/tmp/foo",
	}
	args := dt.getArgs(func(k string) string {
		if k == "no_proxy" {
			return "127.0.0.1"
		}
		return ""
	})
	var (
		image   bool
		noProxy bool
		mem     bool
		dir     bool
	)
	for i := range args {
		if args[i] == "-v" {
			assert.Equal("/tmp/foo:/src", args[i+1])
			dir = true
		}
		if args[i] == "-m" {
			assert.Equal("256m", args[i+1])
			mem = true
		}
		if args[i] == "-e" {
			assert.Equal("no_proxy", args[i+1])
			noProxy = true
		}
		if args[i] == "test" {
			image = true
			assert.Equal("arg1", args[i+1])
		}
	}
	assert.True(image)
	assert.True(noProxy)
	assert.True(mem)
	assert.True(dir)
}
