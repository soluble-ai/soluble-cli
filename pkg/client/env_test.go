package client

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
	os.Setenv("BUILDKITE_COMMAND", "yyy")
	env := getCIEnv()
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
