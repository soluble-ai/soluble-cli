package xcp

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/api"
)

var metadataCommands = map[string]string{
	"SOLUBLE_METADATA_GIT_BRANCH":       "git rev-parse --abbrev-ref HEAD",
	"SOLUBLE_METADATA_GIT_COMMIT":       "git rev-parse HEAD",
	"SOLUBLE_METADATA_GIT_COMMIT_SHORT": "git rev-parse --short HEAD",
	"SOLUBLE_METADATA_GIT_TAG":          "git describe --tags",
	"SOLUBLE_METADATA_GIT_REMOTE":       "git ls-remote --get-url",
}

var (
	// We explicitly exclude a few keys due to their sensitive values.
	// The substrings below will cause the environment variable to be
	// skipped (not recorded).
	substringOmitEnv = []string{
		"SECRET", "KEY", "PRIVATE", "PASSWORD",
		"PASSPHRASE", "CREDS", "TOKEN", "AUTH",
		"ENC", "JWT",
		"_USR", "_PSW", // Jenkins credentials()
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
	}
)

var _ api.Option = WithCIEnv

// Include CI-related environment variables in the request.
func WithCIEnv(req *resty.Request) {
	if req.Method == "GET" {
		req.SetQueryParams(GetCIEnv())
	} else {
		req.SetMultipartFormData(GetCIEnv())
	}
}

// Include CI-related information in the body of a request
func WithCIEnvBody(req *resty.Request) {
	body := jnode.NewObjectNode()
	for k, v := range GetCIEnv() {
		body.Put(k, v)
	}
	req.SetBody(body)
}

// For XCPPost, include a file from a reader.
func WithFileFromReader(param, filename string, reader io.Reader) api.Option {
	return func(req *resty.Request) {
		req.SetFileReader(param, filename, reader)
	}
}

func GetCIEnv() map[string]string {
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
			strings.HasPrefix(k, "CIRCLE_") ||
			strings.HasPrefix(k, "GITLAB_") ||
			strings.HasPrefix(k, "CI_") ||
			strings.HasPrefix(k, "BUILDKITE_") {
			values[k] = v

			// and if we haven't set a CI system yet, set it
			if ciSystem == "" {
				idx := strings.Index(k, "_")
				ciSystem = k[:idx]
			}
		}
	}
	values["SOLUBLE_METADATA_CI_SYSTEM"] = ciSystem

	// evaluate the "easy" metadata commands
	for k, command := range metadataCommands {
		argv := strings.Split(command, " ")
		// #nosec G204
		cmd := exec.Command(argv[0], argv[1:]...)
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
