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

package util

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/soluble-ai/go-jnode"
)

// GetMetadata fetches the runtime information about the CI system and the repo
func GetMetadata() (*jnode.Node, error) {
	// Environment variables
	allEnvs := make(map[string]string)
	for _, e := range os.Environ() {
		split := strings.Split(e, "=")
		allEnvs[split[0]] = split[1]
	}
	// ...but we only want to capture CI environment variables
	envs := make(map[string]string)
	for k, v := range allEnvs {
		if strings.HasPrefix(k, "GITHUB_") ||
			strings.HasPrefix(k, "CIRCLE_") ||
			strings.HasPrefix(k, "GITLAB_") ||
			strings.HasPrefix(k, "CI_") ||
			strings.HasPrefix(k, "BUILDKITE_") {
			envs[k] = v
		}
	}

	// Git remote
	cmd := exec.Command("git", "remote", "-v")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	entries := strings.Split(string(out), "\\n")
	var remotes []string
	for _, e := range entries {
		startIdx := strings.Index(e, "\t")
		endIdx := strings.Index(e, " ")
		if startIdx == -1 || endIdx == -1 {
			continue
		}
		remote := e[startIdx+1 : endIdx]
		remotes = append(remotes, remote)
	}
	remote := remotes[0]

	// Git branch
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err = cmd.Output()
	if err != nil {
		return nil, err
	}
	branch := string(out)[:len(out)-1] // trim newline

	// Hostname
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	metadata := jnode.NewObjectNode().
		Put("gitRemote", remote).
		Put("gitBranch", branch).
		Put("hostname", hostname).
		Put("os", runtime.GOOS)
	return metadata, nil
}
