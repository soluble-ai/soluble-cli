// Copyright 2020 Soluble Inc
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

package api

import (
	"errors"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/soluble-ai/go-jnode"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	assert := assert.New(t)
	c := NewClient(&Config{
		APIServer: "https://api.soluble.cloud",
	})
	c.Organization = "1234"
	httpmock.ActivateNonDefault(c.Client.GetClient())
	t.Cleanup(httpmock.Deactivate)
	httpmock.RegisterResponder("GET", "https://api.soluble.cloud/api/v1/org/1234/foo",
		httpmock.NewJsonResponderOrPanic(http.StatusOK,
			jnode.NewObjectNode().Put("hello", "world")),
	)
	httpmock.RegisterResponder("GET", "https://api.soluble.cloud/api/v1/x/org/1234/foo",
		httpmock.NewJsonResponderOrPanic(http.StatusOK,
			jnode.NewObjectNode().Put("hello", "x world")),
	)
	httpmock.RegisterResponder("GET", "https://api.soluble.cloud/api/v1/download",
		httpmock.NewBytesResponder(http.StatusOK, []byte{1, 2, 3, 4}))
	n, err := c.Get("/api/v1/org/{org}/foo")
	if assert.NoError(err) {
		if !n.IsObject() || n.Path("hello").AsText() != "world" {
			t.Error(n)
		}
	}
	b, err := c.Download("download")
	if assert.NoError(err) {
		assert.Exactly([]byte{1, 2, 3, 4}, b)
	}
	c.APIPrefix = "/api/v1/x"
	n, err = c.Get("org/{org}/foo")
	if assert.NoError(err) {
		if !n.IsObject() || n.Path("hello").AsText() != "x world" {
			t.Error(n)
		}
	}
}

func TestError(t *testing.T) {
	e := httpError("blah")
	if !errors.Is(e, HTTPError) {
		t.Error(e)
	}
}
