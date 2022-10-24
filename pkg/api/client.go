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
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/soluble-ai/go-jnode"
	cfg "github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/version"
)

const (
	orgToken = "{org}"
)

var UserAgent = "soluble-cli/" + version.Version

type Option interface {
	Apply(*resty.Request)
}

type httpError string

var HTTPError httpError

type Client struct {
	*resty.Client
	Config
}

type Config struct {
	Organization     string
	Domain           string
	APIToken         string
	LaceworkAPIToken string
	APIServer        string
	APIPrefix        string
	Debug            bool
	TLSNoVerify      bool
	Timeout          time.Duration
	RetryCount       int
	RetryWaitSeconds float64
	Headers          []string
}

var RClient = resty.New() // exposed for use with httpmock

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
		Client: RClient,
		Config: *config,
	}
	if c.APIPrefix == "" {
		c.APIPrefix = "/api/v1"
	}
	apiServer := config.APIServer
	c.SetBaseURL(apiServer)
	c.SetLogger(logger(0))
	if log.Level == log.Trace {
		c.Client.Debug = true
	}
	if config.TLSNoVerify {
		log.Warnf("{warning:Disabling TLS verification of %s}", apiServer)
		c.SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
		})
	}
	c.SetHeader("User-Agent", UserAgent)
	c.EnableTrace()
	c.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		info := r.Request.TraceInfo()
		log.Tracef("%+v\n", info)
		return nil
	})
	c.AddRetryCondition(func(r *resty.Response, err error) bool {
		if r != nil && r.StatusCode() >= 500 && r.StatusCode() <= 599 {
			return true
		}
		return false
	})
	c.AddRetryHook(func(r *resty.Response, err error) {
		type contextKey int
		const retryInfoKey contextKey = 0
		type retryInfo struct {
			t time.Time
		}
		if r.Request.Attempt == 1 {
			log.Warnf("Retrying {info:%s} {primary:%s}", r.Request.Method, r.Request.URL)
			r.Request.SetContext(context.WithValue(r.Request.Context(), retryInfoKey,
				&retryInfo{t: r.Request.Time}))
		} else {
			info := r.Request.Context().Value(retryInfoKey).(*retryInfo)
			log.Warnf("  {secondary:... retrying attempt %d after %s}", r.Request.Attempt,
				time.Since(info.t).Truncate(time.Millisecond))
			info.t = time.Now()
		}
	})
	c.SetTimeout(config.Timeout)
	c.SetRetryCount(config.RetryCount)
	c.SetRetryMaxWaitTime(16 * time.Second)
	if config.RetryWaitSeconds > 0 {
		c.SetRetryWaitTime(time.Duration(config.RetryWaitSeconds*1000) * time.Millisecond)
	}
	for _, header := range config.Headers {
		nv := strings.Split(header, ":")
		c.SetHeader(nv[0], nv[1])
	}
	return c
}

func (c *Client) ConfigureAuthHeaders(headers http.Header) {
	if c.LaceworkAPIToken != "" {
		headers.Set("X-LW-Domain", c.Domain)
		headers.Set("X-LW-Authorization", fmt.Sprintf("Token %s", c.LaceworkAPIToken))
		log.Debugf("Using lacework authentication")
	} else if c.APIToken != "" {
		headers.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIToken))
	}
}

func (c *Client) execute(r *resty.Request, method, path string, options []Option) error {
	// set r.Method here so that options can do different things
	// depending on the method
	r.Method = method
	for _, opt := range options {
		opt.Apply(r)
	}
	if strings.Contains(path, orgToken) {
		if c.Organization == "" {
			log.Errorf("An organization must be specified with --organization or configuring one with {info:%s configure --organization}",
				cfg.CommandInvocation())
			return fmt.Errorf("organization is required")
		}
		path = strings.ReplaceAll(path, orgToken, c.Organization)
	}
	if c.Organization != "" {
		r.SetHeader("X-SOLUBLE-ORG-ID", c.Organization)
	}
	c.ConfigureAuthHeaders(r.Header)
	if len(path) > 0 && path[0] != '/' {
		path = fmt.Sprintf("%s/%s", c.APIPrefix, path)
	}
	resp, err := r.Execute(method, path)
	switch resp.Header().Get("Content-Type") {
	case "text/plain":
		fallthrough
	case "application/octet-stream":
		// Almost everything we get from api-server is JSON except
		// for assessment files.  So this is a workaround to put
		// the text content in the JSON object.
		content := resp.Body()
		if n, ok := r.Result.(*jnode.Node); ok {
			n.Put("contentBytes", jnode.NewNode(content))
		}
	}
	for _, opt := range options {
		if c, ok := opt.(io.Closer); ok {
			_ = c.Close()
		}
	}
	switch {
	case err != nil:
		return err
	case resp != nil && resp.IsError():
		t := time.Since(r.Time).Truncate(time.Millisecond)
		log.Errorf("{info:%s} {primary:%s} returned {danger:%d} in {secondary:%s}\n", r.Method,
			r.URL, resp.StatusCode(), t)
		log.Errorf("{warning:%s}\n", resp.String())
		if resp.StatusCode() == 401 || resp.StatusCode() == 404 {
			log.Infof("Are you not logged in?  Use {info:%s auth profile} to verify.", cfg.CommandInvocation())
			log.Infof("See {primary:https://docs.lacework.com/iac/} for more information.")
		}
		return httpError(fmt.Sprintf("%s returned %d", r.URL, resp.StatusCode()))
	default:
		t := time.Since(r.Time).Truncate(time.Millisecond)
		log.Tracef("%v", resp.Result())
		log.Infof("{info:%s} {primary:%s} returned {success:%d} in {secondary:%s}\n", r.Method,
			r.URL, resp.StatusCode(), t)
		return nil
	}
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

func (c *Client) XCPPost(module string, files []string, values map[string]string, options ...Option) (*jnode.Node, error) {
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

type optionFunc struct {
	f func(*resty.Request)
}

func (f optionFunc) Apply(req *resty.Request) {
	f.f(req)
}

func OptionFunc(f func(*resty.Request)) Option {
	return optionFunc{f}
}

type closeableOptionFunc struct {
	optionFunc
	close func() error
}

func (c closeableOptionFunc) Close() error {
	return c.close()
}

func CloseableOptionFunc(f func(req *resty.Request), close func() error) Option {
	return closeableOptionFunc{
		optionFunc: optionFunc{f},
		close:      close,
	}
}
