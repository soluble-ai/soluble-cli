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

package download

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestDownload(t *testing.T) {
	setupHTTP()
	m := setupManager()
	if result := m.List(); len(result) != 0 {
		t.Error("non-empty list")
	}
	d, err := m.Install(&Spec{Name: "hello", RequestedVersion: "1.0", URL: "https://example.com/hello.tar.gz"})
	if err != nil {
		t.Fatal(err)
	}
	result := m.List()
	if len(result) != 1 {
		t.Error("after install, wrong # of results", result)
	}
	meta := m.GetMeta(d.Name)
	if meta == nil {
		t.Error("no meta after install")
	}
	_, err = m.Install(&Spec{Name: "hello", RequestedVersion: "2.0", URL: "https://example.com/hello.tar.gz"})
	if err != nil {
		t.Error(err)
	}
	meta = m.GetMeta(d.Name)
	if meta == nil || len(meta.Installed) != 2 {
		t.Error("expecting 2 versions")
	}
	if err = m.Remove("hello", "1.0"); err != nil {
		t.Error(err)
	}
	meta = m.GetMeta(d.Name)
	if meta == nil || len(meta.Installed) != 1 || meta.Installed[0].Version != "2.0" {
		t.Error("expecting only version 2 now")
	}
	if err = m.Remove("hello", ""); err != nil {
		t.Error(err)
	}
	meta = m.GetMeta(d.Name)
	if meta != nil {
		t.Error("expecting hello removed", meta)
	}
}

func TestDownloadZip(t *testing.T) {
	setupHTTP()
	m := setupManager()
	d, err := m.Install(&Spec{Name: "hello-zip", RequestedVersion: "1.0", URL: "https://example.com/hello.zip"})
	if err != nil {
		t.Error(err)
	}
	_, err = os.Stat(filepath.Join(d.Dir, "README.txt"))
	if err != nil {
		t.Error(err)
	}
}

type apiServer string

func (apiServer) GetOrganization() string { return "9999" }
func (apiServer) GetHostURL() string      { return "https://example.com/secure" }
func (a apiServer) GetAuthToken() string  { return string(a) }

func TestAPIServerArtifact(t *testing.T) {
	setupHTTP()
	m := setupManager()
	_, err := m.Install(&Spec{
		Name: "secure", RequestedVersion: "1.0", APIServerArtifact: "/hello.zip",
		APIServer: apiServer(""),
	})
	if err == nil {
		t.Fatal("should have failed")
	}
	_, err = m.Install(&Spec{
		Name: "secure", RequestedVersion: "1.0", APIServerArtifact: "/hello.zip",
		APIServer: apiServer("foo"),
	})
	if err != nil {
		t.Fatal("should have worked")
	}
}

func setupManager() *Manager {
	m := NewManager()
	dir, err := os.MkdirTemp("", "downloadtest*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	m.downloadDir = dir
	return m
}

func setupHTTP() {
	httpmock.Activate()
	registerTestArchive("hello.tar.gz", false)
	registerTestArchive("hello.zip", false)
	registerTestArchive("hello.zip", true)
}

func registerTestArchive(name string, auth bool) {
	dat, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		panic(err)
	}
	r := httpmock.NewBytesResponder(200, dat)
	if auth {
		authr := func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("Authorization") != "Bearer foo" {
				return httpmock.NewStringResponse(403, "Denied"), nil
			}
			return r(req)
		}
		httpmock.RegisterResponder("GET", fmt.Sprintf("https://example.com/secure/%s", name), authr)
	} else {
		httpmock.RegisterResponder("GET", fmt.Sprintf("https://example.com/%s", name), r)
	}
}

func TestNodots(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("__foo_bar___bar.zip", noslashdotdots("../foo/bar/../bar.zip"))
}
