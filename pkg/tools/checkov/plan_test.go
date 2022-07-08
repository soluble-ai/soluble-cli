package checkov

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTerraformNativePlan(t *testing.T) {
	ok, err := isTerraformNativePlan("testdata/main.tfplan")
	assert.NoError(t, err)
	assert.True(t, ok)
}
