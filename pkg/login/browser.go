package login

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/browser"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type BrowserFlow struct {
	appURL          string
	state           string
	endpoint        *callbackEndpoint
	endpointPort    int
	openBrowserFunc func(string) error
	http            *http.Client
}

func NewBrowserFlow(appURL string, state string) Flow {
	return &BrowserFlow{
		appURL:          appURL,
		state:           state,
		endpoint:        &callbackEndpoint{origin: appURL},
		openBrowserFunc: browser.OpenURL,
		http:            &http.Client{},
	}
}

func (f *BrowserFlow) Run() (*Response, error) {
	var err error
	f.endpointPort, err = f.endpoint.listen()
	if err != nil {
		return nil, err
	}
	log.Infof("Started local webserver on port %d", f.endpointPort)
	authURL := fmt.Sprintf("%s/api/v1/auth/cli-login/%s/%d", f.appURL, f.state, f.endpointPort)
	log.Infof("Opening browser URL {primary:%s}", authURL)
	ch := f.endpoint.serveOne()
	if err := f.openBrowserFunc(authURL); err != nil {
		return nil, err
	}
	var n *jnode.Node
	select {
	case n = <-ch:
	case <-time.After(time.Minute):
		return nil, fmt.Errorf("authentication timed out")
	}
	if !n.IsObject() || n.Path("code").AsText() == "" {
		log.Infof("{info:%s} did not return a successful authentication code", f.appURL)
		return nil, fmt.Errorf("no response from servr")
	}
	tokenURL := fmt.Sprintf("%s/api/v1/auth/cli-login-code", f.appURL)
	log.Infof("Getting authentication token from {primary:%s}", tokenURL)
	resp, err := f.http.PostForm(tokenURL, url.Values{
		"state": []string{f.state},
		"code":  []string{n.Path("code").AsText()},
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to server: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}
	dat, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("garbled response: %w", err)
	}
	result := Response{}
	if err := json.Unmarshal(dat, &result); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	result.defaultAPIServer(f.appURL)
	return &result, nil
}

func (f *BrowserFlow) Close() error {
	return util.CloseAll(f.endpoint)
}
