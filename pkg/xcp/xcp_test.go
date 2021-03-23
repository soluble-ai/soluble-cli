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

package xcp

import (
	"os"
	"strings"
	"testing"
)

func TestGetCIEnv(t *testing.T) {
	// xxx must not be included, yyy must be included
	os.Setenv("PASSWORD", "xxx")
	os.Setenv("GITHUB_TOKEN", "xxx")
	os.Setenv("GITHUB_BRANCH", "yyy")
	os.Setenv("BUILDKITE_AGENT_ACCESS_TOKEN", "xxx")
	os.Setenv("BUILDKITE_COMMAND", "xxx")
	os.Setenv("BUILDKITE_S3_ACCESS_URL", "xxx")
	env := GetCIEnv(".")
	for k, v := range env {
		if v == "xxx" {
			t.Error(k, v)
		}
	}
	for _, kv := range os.Environ() {
		if strings.HasSuffix(kv, "=yyy") {
			if env[kv[0:len(kv)-4]] != "yyy" {
				t.Error(kv)
			}
		}
	}
}

func TestNormalizeGitRemote(t *testing.T) {
	if s := normalizeGitRemote("git@github.com:fizz/buzz.git"); s != "github.com/fizz/buzz" {
		t.Error(s)
	}
}
