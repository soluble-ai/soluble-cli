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
	"github.com/stretchr/testify/assert"
)

func TestGetCIEnv(t *testing.T) {
	assert := assert.New(t)
	saveEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, kv := range saveEnv {
			eq := strings.Index(kv, "=")
			os.Setenv(kv[0:eq], kv[eq+1:])
		}
	}()
	// xxx must not be included, yyy must be included
	os.Setenv("PASSWORD", "xxx")
	os.Setenv("GITHUB_TOKEN", "xxx")
	os.Setenv("GITHUB_BRANCH", "yyy")
	os.Setenv("BUILDKITE_AGENT_ACCESS_TOKEN", "xxx")
	os.Setenv("BUILDKITE_COMMAND", "xxx")
	os.Setenv("BUILDKITE_S3_ACCESS_URL", "xxx")
	os.Setenv("BITBUCKET_BUILD_NUMBER", "yyy")
	os.Setenv("BITBUCKET_STEP_OIDC_TOKEN", "xxx")
	os.Setenv("ATLANTIS_TERRAFORM_VERSION", "yyy")
	os.Setenv("PULL_NUM", "yyy")
	os.Setenv("PULL_AUTHOR", "yyy")
	os.Setenv("REPO_REL_DIR", "yyy")
	env := GetCIEnv(".")
	for k, v := range env {
		if v == "xxx" {
			t.Error(k, v)
		}
	}

	// make sure atlantis env variables are available
	assert.True(contains(env, "ATLANTIS_TERRAFORM_VERSION"))
	assert.True(contains(env, "PULL_NUM"))
	assert.True(contains(env, "PULL_AUTHOR"))

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

func contains(s map[string]string, searchStr string) bool {
	for k, _ := range s {
		if k == searchStr {
			return true
		}
	}
	return false
}
