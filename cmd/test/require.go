package test

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/config"
)

func RequireAPIToken(t *testing.T) {
	t.Helper()
	if !HaveAPIToken() {
		t.Skip("test requires authentication")
	}
}

func HaveAPIToken() bool {
	config.Load()
	return config.Config.APIToken != ""
}
