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
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/version"
)

const (
	orgToken = "{org}"
)

type Config struct {
	Organization     string
	APIToken         string
	APIServer        string
	APIPrefix        string
	Debug            bool
	TLSNoVerify      bool
	TimeoutSeconds   int
	RetryCount       int
	RetryWaitSeconds float64
	Headers          []string
}

type Option func(*resty.Request)

type httpError string

var HTTPError httpError

type Client struct {
	*resty.Client
	Config
}

func (h httpError) Error() string {
	return string(h)
}

func (h httpError) Is(err error) bool {
	switch err.(type) {
	case httpError:
		return true
	default:
		return false
	}
}

func NewClient(config *Config) *Client {
	c := &Client{
		Client: resty.New(),
		Config: *config,
	}
	c.Token = config.APIToken
	if c.APIPrefix == "" {
		c.APIPrefix = "/api/v1"
	}

	apiServer := config.APIServer
	c.SetBaseURL(apiServer)
	if log.Level == log.Debug {
		c.Client.Debug = true
	}
	if config.TLSNoVerify {
		log.Warnf("{warning:Disabling TLS verification of %s}", apiServer)
		c.SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
		})
	}
	c.SetHeader("User-Agent", "soluble-cli/"+version.Version)
	c.EnableTrace()
	c.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		info := r.Request.TraceInfo()
		log.Debugf("{warning:%+v}\n", info)
		return nil
	})

	c.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		t := r.Request.TraceInfo().TotalTime.Truncate(time.Millisecond)
		if r.IsError() {
			log.Errorf("{info:%s} {primary:%s} returned {danger:%d} in {secondary:%s}\n", r.Request.Method,
				r.Request.URL, r.StatusCode(), t)
			log.Errorf("{warning:%s}\n", r.String())
			if r.StatusCode() == 401 {
				log.Infof("Are you not logged in?  Use {primary:soluble login} to login, or {primary:soluble auth profile} to verify")
			}
			return httpError(fmt.Sprintf("%s returned %d", r.Request.URL, r.StatusCode()))
		}
		log.Debugf("%v", r.Result())
		log.Infof("{info:%s} {primary:%s} returned {success:%d} in {secondary:%s}\n", r.Request.Method,
			r.Request.URL, r.StatusCode(), t)
		return nil
	})
	c.SetTimeout(time.Duration(config.TimeoutSeconds) * time.Second)
	c.SetRetryCount(config.RetryCount)
	if config.RetryWaitSeconds > 0 {
		c.SetRetryWaitTime(time.Duration(config.RetryWaitSeconds*1000) * time.Millisecond)
	}
	for _, header := range config.Headers {
		nv := strings.Split(header, ":")
		c.SetHeader(nv[0], nv[1])
	}
	return c
}

func (c *Client) execute(r *resty.Request, method, path string, options []Option) error {
	// set r.Method here so that options can do different things
	// depending on the method
	r.Method = method
	for _, opt := range options {
		opt(r)
	}
	if strings.Contains(path, orgToken) {
		if c.Organization == "" {
			log.Errorf("An organization must be specified with --organization or configuring one with `cli-config set organization <org-id>`")
			return fmt.Errorf("organization is required")
		}
		path = strings.ReplaceAll(path, orgToken, c.Organization)
	}
	if len(path) > 0 && path[0] != '/' {
		path = fmt.Sprintf("%s/%s", c.APIPrefix, path)
	}
	_, err := r.Execute(method, path)
	return err
}

func (c *Client) Post(path string, body *jnode.Node, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()
	if err := c.execute(c.R().SetBody(body).SetResult(result), resty.MethodPost, path, options); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) Get(path string, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()
	if err := c.execute(c.R().SetResult(result), resty.MethodGet, path, options); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetWithParams(path string, params map[string]string, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()

	if err := c.execute(c.R().SetQueryParams(params).SetResult(result), resty.MethodGet, path, options); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) Delete(path string, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()
	if err := c.execute(c.R().SetResult(result), resty.MethodDelete, path, options); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) Patch(path string, body *jnode.Node, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()
	if err := c.execute(c.R().SetResult(result).SetBody(body), resty.MethodPatch, path, options); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetClient() *resty.Client {
	return c.Client
}

func (c *Client) XCPPost(orgID string, module string, files []string, values map[string]string, options ...Option) (*jnode.Node, error) {
	if module == "" {
		return nil, fmt.Errorf("module parameter is required")
	}
	req := c.R()
	for i, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		req.SetFileReader(fmt.Sprintf("file_%d", i), filepath.Base(file), f)
	}
	req.SetHeader("X-SOLUBLE-ORG-ID", orgID)
	req.SetMultipartFormData(values)
	result := jnode.NewObjectNode()
	req.SetResult(result)
	if err := c.execute(req, resty.MethodPost, fmt.Sprintf("/api/v1/xcp/%s/data", module), options); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetOrganization() string {
	return c.Organization
}

func (c *Client) GetHostURL() string {
	return c.HostURL
}

func (c *Client) GetAuthToken() string {
	return c.APIToken
}
