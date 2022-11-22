package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAppURL(t *testing.T) {
	t.Setenv("SOLUBLE_API_SERVER", "") // github action sets this, so make sure it's not
	c := &Config{
		APIServer: "https://api.example.com",
	}
	if u := c.GetAppURL(); u != "https://app.example.com" {
		t.Error(c.APIServer, u)
	}
}

func TestGetDomain(t *testing.T) {
	c := &Config{LaceworkAccount: "qan"}
	assert.Equal(t, "qan.lacework.net", c.GetDomain())
}
