package test

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/config"
)

func RequireAPIToken(t *testing.T) {
	t.Helper()
	config.Load()
	if config.Get().APIToken == "" {
		t.Skip("test requires authentication")
	}
}
