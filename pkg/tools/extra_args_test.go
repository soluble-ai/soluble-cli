package tools

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestExtraArgs(t *testing.T) {
	assert := assert.New(t)
	var ex ExtraArgs
	c := &cobra.Command{
		Args: ex.ArgsValue(),
		Run:  func(cmd *cobra.Command, args []string) {},
	}
	c.SetArgs([]string{"hello", "world"})
	assert.NoError(c.Execute())
	assert.Equal([]string{"hello", "world"}, []string(ex))
}
