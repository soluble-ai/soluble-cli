package api

import (
	"os"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestSaveCredentials(t *testing.T) {
	assert := assert.New(t)
	dir, err := os.MkdirTemp("", "cli*")
	if !assert.NoError(err) {
		return
	}
	t.Setenv("SOLUBLE_CONFIG_DIR", dir)
	creds := loadCredentials()
	assert.NotNil(creds)
	creds["default"] = &ProfileCredentials{
		Token:     "12345",
		ExpiresAt: time.Now(),
	}
	assert.NoError(creds.Save())
	credentials = nil
	creds = loadCredentials()
	pd := creds.Find("default")
	if assert.NotNil(pd) {
		assert.Equal("12345", pd.Token)
	}
	pd = creds.Find("test")
	pd.Token = "678910"
	assert.NoError(creds.Save())
	credentials = nil
	creds = loadCredentials()
	pt := creds.Find("test")
	if assert.NotNil(pt) {
		assert.Equal("678910", pt.Token)
	}
}

func TestLoadCredentials(t *testing.T) {
	assert := assert.New(t)
	t.Setenv("SOLUBLE_CONFIG_DIR", "testdata")
	creds := loadCredentials()
	if assert.NotNil(creds["default"]) {
		assert.Equal("12345", creds["default"].Token)
	}
}

func TestRefresh(t *testing.T) {
	assert := assert.New(t)
	httpmock.Activate()
	t.Cleanup(httpmock.Deactivate)
	dir, err := os.MkdirTemp("", "cli*")
	if !assert.NoError(err) {
		return
	}
	t.Setenv("SOLUBLE_CONFIG_DIR", dir)
	httpmock.RegisterResponder("POST", "https://test.lacework.net/api/v2/access/tokens",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"token":     "12345",
			"expiresAt": time.Now().Add(5 * time.Minute),
		}))
	creds := loadCredentials()
	dc := creds.Find("default")
	assert.NoError(dc.RefreshToken("test.lacework.net", "TEST_12345", "_12345"))
	assert.Equal("12345", dc.Token)
	assert.NoError(creds.Save())
	credentials = nil
	creds = loadCredentials()
	dc = creds.Find("default")
	assert.Equal("12345", dc.Token)
}
