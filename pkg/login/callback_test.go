package login

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
)

func TestListen(t *testing.T) {
	e := &callbackEndpoint{}
	defer e.Close()
	_, err := e.listen()
	if err != nil {
		t.Fatal(err)
	}
}

func TestListSkipPorts(t *testing.T) {
	_, _ = net.Listen("tcp", "127.0.0.1:8085")
	e := &callbackEndpoint{}
	defer e.Close()
	port, err := e.listen()
	if err != nil {
		t.Fatal(err)
	}
	if port == 8085 {
		t.Error(port)
	}
	e.portSearchEnd = 8086
	_, err = e.listen()
	if err == nil {
		t.Fatal("port should not have been found")
	}
}

func TestHandle(t *testing.T) {
	e := &callbackEndpoint{}
	defer e.Close()
	port, err := e.listen()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		url := fmt.Sprintf("http://127.0.0.1:%d/auth/callback", port)
		opt, err := http.NewRequest("OPTIONS", url, nil)
		if err != nil {
			panic(err)
		}
		optr, err := http.DefaultClient.Do(opt)
		if err != nil {
			panic(err)
		}
		if optr.StatusCode != 200 {
			panic(optr)
		}
		r := strings.NewReader(`{"code":"foo"}`)
		// #nosec G107
		_, _ = http.Post(url, "application/json", r)
	}()
	n := <-e.serveOne()
	if err != nil {
		t.Fatal(err)
	}
	if n == nil || n.Path("code").AsText() != "foo" {
		t.Error(n)
	}
}
