package test

import (
	"strings"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/config"
)

func RequireAPIToken(t *testing.T) {
	t.Helper()
	config.Load()
	if config.Config.APIToken == "" {
		t.Skip("test requires authentication")
	}
	if !strings.HasSuffix(config.Config.ProfileName, "-test") {
		t.Log("Integration testing requires running with a profile that ends with -test")
		t.Log("(You can copy an existing profile with \"... config new-profile --name demo-test --copy-from demo\")")
		t.FailNow()
	}
}
