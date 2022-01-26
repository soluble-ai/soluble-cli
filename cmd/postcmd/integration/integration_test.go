//go:build integration

package integration

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
)

func TestPost(t *testing.T) {
	test.RequireAPIToken(t)
	cmd := test.NewCommand(t, "post", "-m", "test", "-p", "hello=world")
	cmd.Must(cmd.Run())
}
