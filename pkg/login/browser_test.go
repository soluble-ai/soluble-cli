package login

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestBrowserFlow(t *testing.T) {
	f := NewBrowserFlow("https://app.example.com", "xxx").(*BrowserFlow)
	defer f.Close()
	httpmock.ActivateNonDefault(f.http)
	httpmock.RegisterResponder("POST", "https://app.example.com/api/v1/auth/cli-login-code",
		httpmock.NewStringResponder(200, `{"userId":"u-1234","orgId":"9000","token":"foo"}`))
	f.openBrowserFunc = func(s string) error {
		go func() {
			rd := strings.NewReader(`{"code":"xxx"}`)
			r, err := http.Post(fmt.Sprintf("http://localhost:%d/auth/callback", f.endpointPort), "application/json", rd)
			if err != nil {
				panic(err)
			}
			if r == nil || r.StatusCode != 200 {
				panic(r)
			}
		}()
		return nil
	}
	r, err := f.Run()
	if err != nil {
		t.Error(err)
	}
	if r == nil || r.Token != "foo" || r.UserID != "u-1234" || r.OrgID != "9000" || r.APIServer != "https://api.example.com" {
		t.Fatal(r)
	}
}
