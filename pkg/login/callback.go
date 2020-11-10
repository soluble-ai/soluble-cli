package login

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type callbackEndpoint struct {
	origin          string
	portSearchStart int
	portSearchEnd   int
	s               *http.Server
	r               chan *jnode.Node
	ln              net.Listener
}

func (e *callbackEndpoint) listen() (int, error) {
	if e.portSearchStart == 0 {
		e.portSearchStart = 8085
	}
	if e.portSearchEnd == 0 {
		e.portSearchEnd = 8185
	}
	for port := e.portSearchStart; port < e.portSearchEnd; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			e.ln = ln
			return port, nil
		}
	}
	log.Errorf(`Cannot start a local webserver listening on any port between %d and %d.
Check your firewall settings or locally running programs that may be blocking
or using those ports.	
`, e.portSearchStart, e.portSearchEnd)
	return 0, fmt.Errorf("cannot find an available port in the range %d - %d to listen on", e.portSearchStart, e.portSearchEnd)
}

func (e *callbackEndpoint) serveOne() <-chan *jnode.Node {
	e.r = make(chan *jnode.Node)
	e.s = &http.Server{
		Handler: http.HandlerFunc(e.handleAuth),
	}
	go func() {
		err := e.s.Serve(e.ln)
		if !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("Local webserver unexpectedly didn't start: %s", err.Error())
		}
	}()
	return e.r
}

func (e *callbackEndpoint) handleAuth(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/auth/callback" || !(req.Method == "POST" || req.Method == "OPTIONS") {
		w.WriteHeader(400)
		log.Errorf("Received invalid request on local webserver %s %s", req.Method, req.URL.Path)
		e.r <- nil
		return
	}
	w.Header().Add("Access-Control-Allow-Origin", e.origin)
	w.Header().Add("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Add("Access-Control-Allow-Headers", "*")
	w.Header().Add("Access-Control-Max-Age", "600")
	w.Header().Add("Content-type", "text/plain")
	if req.Method == "POST" {
		dat, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Errorf("Could not read oauth response: {danger:%s}", err.Error())
			e.r <- nil
			return
		}
		n, err := jnode.FromJSON(dat)
		if err != nil {
			log.Errorf("Invalid oauth response: {danger:%s}", err.Error())
		}
		log.Infof("Received authentication code from server")
		e.r <- n
	}
	_, _ = w.Write([]byte("OK"))
}

func (e *callbackEndpoint) Close() error {
	if e.s == nil {
		if e.ln != nil {
			return e.ln.Close()
		}
		return nil
	} else {
		return e.s.Close()
	}
}
