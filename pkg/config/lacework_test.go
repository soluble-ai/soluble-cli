package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLaceworkProfile(t *testing.T) {
	LoadLaceworkProfiles("testdata/lacework.toml")
	t.Cleanup(func() { LoadLaceworkProfiles("") })
	p := getLaceworkProfile("default")
	assert := assert.New(t)
	if assert.NotNil(p) {
		assert.Equal("default", p.Name)
		assert.Equal("test1", p.Account)
		assert.Equal("TEST1_10E74F8519784222B2E5ACCC1D297478E7E2848442EF6E6", p.APIKey)
		assert.Equal("_8c80ac4717ac79cac23dea7edf5bcbbb", p.APISecret)
	}
	p = GetDefaultLaceworkProfile()
	if assert.NotNil(p) {
		assert.Equal("default", p.Name)
	}
	p = getLaceworkProfile("test2")
	if assert.NotNil(p) {
		assert.Equal("test2", p.Name)
		assert.Equal("test2.corp", p.Account)
	}
}
