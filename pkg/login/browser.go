package login

import (
	"fmt"
	"time"

	"github.com/pkg/browser"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type BrowserLeg struct {
	openBrowserFunc func(url string) error
}

func (f *BrowserLeg) startEndpoint(appURL string) (*callbackEndpoint, int, error) {
	endpoint := &callbackEndpoint{origin: appURL}
	endpointPort, err := endpoint.listen()
	return endpoint, endpointPort, err
}

func (f *BrowserLeg) GetCode(appURL string, state string) (string, error) {
	endpoint, endpointPort, err := f.startEndpoint(appURL)
	defer func() { _ = endpoint.Close() }()
	if err != nil {
		return "", err
	}
	log.Infof("Started local webserver on port %d", endpointPort)
	authURL := fmt.Sprintf("%s/api/v1/auth/cli-login/%s/%d", appURL, state, endpointPort)
	log.Infof("Opening browser URL {primary:%s}", authURL)
	ch := endpoint.serveOne()
	openBrowserFunc := f.openBrowserFunc
	if openBrowserFunc == nil {
		openBrowserFunc = browser.OpenURL
	}
	if err := openBrowserFunc(authURL); err != nil {
		log.Warnf("No browser? Try running again in headless mode: {warning:soluble login --headless}")
		return "", err
	}
	var n *jnode.Node
	select {
	case n = <-ch:
	case <-time.After(time.Minute):
		return "", fmt.Errorf("authentication timed out")
	}
	if !n.IsObject() || n.Path("code").AsText() == "" {
		log.Infof("{info:%s} did not return a successful authentication code", appURL)
		return "", fmt.Errorf("no response from servr")
	}
	return n.Path("code").AsText(), nil
}
