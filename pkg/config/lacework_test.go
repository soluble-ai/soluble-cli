package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLaceworkProfile(t *testing.T) {
	laceworkConfigFile = "testdata/lacework.toml"
	defer func() { laceworkConfigFile = "" }()
	laceworkProfiles = nil
	p := getLaceworkProfile("", "")
	assert := assert.New(t)
	if assert.NotNil(p) {
		assert.Equal("test1", p.Account)
		assert.Equal("TEST1_10E74F8519784222B2E5ACCC1D297478E7E2848442EF6E6", p.APIKey)
		assert.Equal("_8c80ac4717ac79cac23dea7edf5bcbbb", p.APISecret)
	}
	p = getLaceworkProfile("", "test2")
	if assert.NotNil(p) {
		assert.Equal("test2.corp", p.Account)
	}
}
