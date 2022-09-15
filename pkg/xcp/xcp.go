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
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

var metadataCommands = map[string]string{
	"SOLUBLE_METADATA_GIT_BRANCH":       "git rev-parse --abbrev-ref HEAD",
	"SOLUBLE_METADATA_GIT_COMMIT":       "git rev-parse HEAD",
	"SOLUBLE_METADATA_GIT_COMMIT_SHORT": "git rev-parse --short HEAD",
	"SOLUBLE_METADATA_GIT_DESCRIBE":     "git describe --tags --always",
	"SOLUBLE_METADATA_GIT_REMOTE":       "git ls-remote --get-url",
}

var (
	// We explicitly exclude a few keys due to their sensitive values.
	// The substrings below will cause the environment variable to be
	// skipped (not recorded).
	substringOmitEnv = []string{
		"SECRET", "KEY", "PRIVATE", "PASSWORD",
		"PASSPHRASE", "CREDS", "TOKEN", "AUTH",
		"ENC", "JWT", "password", "username",
		"CLIENT", "TENANT", "_USR", "_PSW", // Jenkins credentials()
	}

	// While we perform the redactions based on substrings above,
	// we also maintain a list of known-sensitive keys to ensure
	// that we never capture these. Unlike above, these are an
	// exact match and not a substring match.
	explicitOmitEnv = []string{
		"BUILDKITE_S3_SECRET_ACCESS_KEY", // Buildkite
		"BUILDKITE_S3_ACCESS_KEY_ID",     // Buildkite
		"BUILDKITE_S3_ACCESS_URL",        // Buildkite
		"BUILDKITE_COMMAND",              // Buildkite
		"BUILDKITE_SCRIPT_PATH",          // Buildkite
		"KEY",                            // CircleCI encrypted-files decryption key
		"CI_DEPLOY_PASSWORD",             // Gitlab
		"CI_DEPLOY_USER",                 // Gitlab
		"CI_JOB_TOKEN",                   // Gitlab
		"CI_JOB_JWT",                     // Gitlab
		"CI_REGISTRY_USER",               // Gitlab
		"CI_REGISTRY_PASSWORD",           // Gitlab
		"CI_REGISTRY_USER",               // Gitlab
		"BITBUCKET_STEP_OIDC_TOKEN",      // Bitbucket
	}
)

// Include CI-related environment variables in the request.
func WithCIEnv(dir string) api.Option {
	return api.OptionFunc(func(req *resty.Request) {
		if req.Method == "GET" {
			req.SetQueryParams(GetCIEnv(dir))
		} else {
			req.SetMultipartFormData(GetCIEnv(dir))
		}
	})
}

// Include CI-related information in the body of a request
func WithCIEnvBody(dir string) api.Option {
	return api.OptionFunc(func(r *resty.Request) {
		body := jnode.NewObjectNode()
		for k, v := range GetCIEnv(dir) {
			body.Put(k, v)
		}
		r.SetBody(body)
	})
}

// For XCPPost, include a file from a reader.
func WithFileFromReader(param, filename string, reader io.Reader) api.Option {
	closer, _ := reader.(io.Closer)
	return api.CloseableOptionFunc(func(req *resty.Request) {
		req.SetFileReader(param, filename, reader)
		log.Debugf("...including {secondary:%s}", filename)
	}, func() error {
		if closer != nil {
			return closer.Close()
		}
		return nil
	})
}

func WithFile(path string) (api.Option, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	filename := filepath.Base(path)
	return WithFileFromReader(filename, filename, f), nil
}

func getCIEnvValues() map[string]string {
	values := map[string]string{}
	allEnvs := make(map[string]string)
	for _, e := range os.Environ() {
		split := strings.Split(e, "=")
		allEnvs[split[0]] = split[1]
	}
	var ciSystem string
	// We don't want all of the environment variables, however.
envLoop:
	for k, v := range allEnvs {
		k = strings.ToUpper(k)
		for _, s := range substringOmitEnv {
			if strings.Contains(k, s) {
				continue envLoop
			}
		}
		for _, s := range explicitOmitEnv {
			if k == s {
				continue envLoop
			}
		}

		// If the key has made it through the filtering above and is
		// from a CI system, we include it.
		if strings.HasPrefix(k, "GITHUB_") ||
			strings.HasPrefix(k, "GIT_PR_") || // iacbot
			strings.HasPrefix(k, "CIRCLE_") ||
			strings.HasPrefix(k, "GITLAB_") ||
			strings.HasPrefix(k, "BUILDKITE_") ||
			strings.HasPrefix(k, "ZODIAC_") ||
			strings.HasPrefix(k, "BITBUCKET_") ||
			strings.HasPrefix(k, "ATLANTIS_") ||
			strings.HasPrefix(k, "JENKINS_") {
			values[k] = v

			// and if we haven't set a CI system yet, set it
			if ciSystem == "" {
				idx := strings.Index(k, "_")
				if idx > 0 {
					ciSystem = k[:idx]
				}
			}
		}

		// for some ci/cd tools such as jenkins some times we don't capture the
		// right env variables based on how it is setup, so let's do best
		// guessing on standard naming conventions and pass those as well
		if strings.HasPrefix(k, "JOB_") ||
			strings.HasPrefix(k, "BUILD_") ||
			strings.HasPrefix(k, "STAGE_") ||
			strings.HasPrefix(k, "RUN_") ||
			strings.HasPrefix(k, "CI_") ||
			strings.HasPrefix(k, "HUDSON_") ||
			strings.HasPrefix(k, "WORKSPACE") ||
			strings.HasPrefix(k, "KUBERNETES_") {
			values[k] = v
		}

		// for atlantis they do not have a prefix in the key for most of them
		// https://www.runatlantis.io/docs/custom-workflows.html#reference
		if strings.EqualFold(k, "DIR") ||
			strings.EqualFold(k, "WORKSPACE") ||
			strings.EqualFold(k, "PULL_NUM") ||
			strings.EqualFold(k, "PULL_AUTHOR") ||
			strings.EqualFold(k, "PROJECT_NAME") ||
			strings.EqualFold(k, "REPO_REL_DIR") {
			values["ATLANTIS_"+k] = v
		}
	}
	values["SOLUBLE_METADATA_CI_SYSTEM"] = ciSystem
	return values
}

func GetCIEnv(dir string) map[string]string {
	values := getCIEnvValues()
	dir = filepath.Clean(dir)
	// evaluate the "easy" metadata commands
	for k, command := range metadataCommands {
		argv := strings.Split(command, " ")
		// #nosec G204
		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.Dir = dir
		out, err := cmd.Output()
		if err == nil {
			values[k] = strings.TrimSpace(string(out))
		}
	}
	if s := normalizeGitRemote(values["SOLUBLE_METADATA_GIT_REMOTE"]); s != "" {
		values["SOLUBLE_METADATA_GIT_REMOTE"] = s
	}

	// Hostname
	h, err := os.Hostname()
	if err == nil {
		values["SOLUBLE_METADATA_HOSTNAME"] = h
	}

	return values
}

func normalizeGitRemote(s string) string {
	// transform "git@github.com:fizz/buzz.git" to "github.com/fizz/buzz"
	at := strings.Index(s, "@")
	dotgit := strings.LastIndex(s, ".git")
	if at > 0 && dotgit > 0 {
		return strings.Replace(s[at+1:dotgit], ":", "/", 1)
	}
	return s
}

var ciSystem *string

func GetCISystem() string {
	if ciSystem == nil {
		val := getCIEnvValues()["SOLUBLE_METADATA_CI_SYSTEM"]
		ciSystem = &val
	}
	return *ciSystem
}
