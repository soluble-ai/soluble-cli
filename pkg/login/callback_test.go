// Copyright 2021 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	defer func() {
		e.portSearchEnd = 8186
	}()
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
