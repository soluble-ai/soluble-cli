//go:build integration

package integration

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
	"github.com/stretchr/testify/assert"
)

func TestConfigShow(t *testing.T) {
	assert := assert.New(t)
	cmd := test.NewCommand(t, "configure", "show")
	cmd.Must(cmd.Run())
	assert.NotEmpty(cmd.Out.Bytes())
}
