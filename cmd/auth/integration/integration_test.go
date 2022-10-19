//go:build integration

package integration

import (
	"fmt"
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
	"github.com/stretchr/testify/assert"
)

func TestAuthProfile(t *testing.T) {
	test.RequireAPIToken(t)
	cmd := test.NewCommand(t, "auth", "profile", "--format", "json")
	cmd.Must(cmd.Run())
	n := cmd.JSON()
	fmt.Println(n)
	assert := assert.New(t)
	assert.NotNil(n)
	assert.NotEmpty(n.Path("data").Path("currentOrgId"), n)
	assert.GreaterOrEqual(n.Path("data").Path("organizations").Size(), 1)
}
