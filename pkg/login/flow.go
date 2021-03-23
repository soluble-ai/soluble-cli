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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type Response struct {
	UserID    string `json:"userId"`
	OrgID     string `json:"orgId"`
	Token     string `json:"token"`
	APIServer string `json:"apiServer"`
}

type Flow struct {
	authCodeLeg AuthCodeLeg
	appURL      string
	state       string
	http        *http.Client
}

type AuthCodeLeg interface {
	GetCode(appURL, state string) (string, error)
}

func NewFlow(appURL string, headless bool) *Flow {
	flow := &Flow{
		state:  MakeState(),
		appURL: appURL,
		http:   &http.Client{},
	}
	if headless {
		flow.authCodeLeg = &HeadlessLeg{}
	} else {
		flow.authCodeLeg = &BrowserLeg{}
	}
	return flow
}

func (f *Flow) Run() (*Response, error) {
	code, err := f.authCodeLeg.GetCode(f.appURL, f.state)
	if err != nil {
		return nil, err
	}
	tokenURL := fmt.Sprintf("%s/api/v1/auth/cli-login-code", f.appURL)
	log.Infof("Getting authentication token from {primary:%s}", tokenURL)
	resp, err := f.http.PostForm(tokenURL, url.Values{
		"state": []string{f.state},
		"code":  []string{code},
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

func (r *Response) defaultAPIServer(appURL string) {
	const httpsAPI = "https://app."
	if strings.HasPrefix(appURL, httpsAPI) {
		r.APIServer = "https://api." + appURL[len(httpsAPI):]
	}
}
