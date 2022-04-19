//go:build integration

package integration

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
	"github.com/stretchr/testify/assert"
)

func TestInventoryCommand(t *testing.T) {
	test.RequireAPIToken(t)
	inv := test.NewCommand(t, "inventory", "--format", "json", "--no-color", "-d", "../../..")
	inv.Must(inv.Run())
	n := inv.JSON()
	assert.Greater(t, n.Path("terraform_modules").Size(), 0)
}
