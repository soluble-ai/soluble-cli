package api

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	assert := assert.New(t)
	c := NewClient(&Config{
		APIServer:        "https://api.soluble.cloud",
		RetryCount:       3,
		RetryWaitSeconds: 0.1,
	})
	c.Organization = "1234"
	httpmock.ActivateNonDefault(c.GetClient().GetClient())
	httpmock.ZeroCallCounters()
	t.Cleanup(httpmock.Deactivate)
	count := 2
	httpmock.RegisterResponder("GET", "https://api.soluble.cloud/api/v1/org/1234/foo",
		func(r *http.Request) (*http.Response, error) {
			count--
			if count == 0 {
				return httpmock.NewJsonResponse(200, map[string]interface{}{"message": "ok"})
			}
			return httpmock.NewJsonResponse(502, map[string]interface{}{
				"message": "not working",
			})
		})
	n, err := c.Get("org/1234/foo")
	assert.NoError(err)
	assert.NotNil(n)
	assert.Equal(2, httpmock.GetTotalCallCount())
}
