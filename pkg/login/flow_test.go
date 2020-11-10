package login

import "testing"

func TestDefaultAPIServer(t *testing.T) {
	r := &Response{}
	r.defaultAPIServer("https://app.soluble.cloud")
	if r.APIServer != "https://api.soluble.cloud" {
		t.Error(r)
	}
}
