package terraformsettings

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	assert := assert.New(t)
	settings := &TerraformSettings{}
	src, err := os.ReadFile("testdata/provider.tf")
	assert.NoError(err)
	Parse("provider.tf", src, settings)
	assert.NotNil(settings.RequiredVersion)
	if settings.RequiredVersion == nil {
		return
	}
	assert.Equal("~> 0.12.6 ", *settings.RequiredVersion)
	assert.Equal("0.12.31", settings.GetTerraformVersion())
}
