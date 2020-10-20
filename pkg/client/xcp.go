package client

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	metaGitRemote string = "SOLUBLE_METADATA_GIT_REMOTE"
	metaGitBranch string = "SOLUBLE_METADATA_GIT_BRANCH"
	metaHostname  string = "SOLUBLE_METADATA_HOSTNAME"
)

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
		"BUILDKITE_S3_SECRET_ACCESS_KEY",
		"BUILDKITE_S3_ACCESS_KEY_ID",
		"BUILDKITE_S3_ACCESS_URL",
		"KEY",                  // CircleCI encrypted-files decryption key
		"CI_DEPLOY_PASSWORD",   // Gitlab
		"CI_DEPLOY_USER",       // Gitlab
		"CI_JOB_TOKEN",         // Gitlab
		"CI_JOB_JWT",           // Gitlab
		"CI_REGISTRY_USER",     // Gitlab
		"CI_REGISTRY_PASSWORD", // Gitlab
		"CI_REGISTRY_USER",     // Gitlab
	}
)

var _ Option = XCPWithCIEnv

// For XCPPost, Include CI-related environment variables in the request
func XCPWithCIEnv(req *resty.Request) {
	req.SetMultipartFormData(getCIEnv())
}

// For XCPPost, include a file from a reader
func XCPWithReader(param, filename string, reader io.Reader) Option {
	return func(req *resty.Request) {
		req.SetFileReader(param, filename, reader)
	}
}

func getCIEnv() map[string]string {
	values := map[string]string{}
	allEnvs := make(map[string]string)
	for _, e := range os.Environ() {
		split := strings.Split(e, "=")
		allEnvs[split[0]] = split[1]
	}
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
		}
	}

	// Git remote
	cmd := exec.Command("git", "remote", "-v")
	out, err := cmd.Output()
	if err == nil {
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
		values[metaGitRemote] = remotes[0]
	}

	// Git Branch
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err = cmd.Output()
	if err == nil {
		values[metaGitBranch] = strings.TrimSpace(string(out))
	}

	// Hostname
	h, err := os.Hostname()
	if err == nil {
		values[metaHostname] = h
	}

	return values
}
