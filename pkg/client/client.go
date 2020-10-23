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

package client

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

type Interface interface {
	Post(path string, body *jnode.Node, options ...Option) (*jnode.Node, error)
	Patch(path string, body *jnode.Node, options ...Option) (*jnode.Node, error)
	Get(path string, options ...Option) (*jnode.Node, error)
	GetWithParams(path string, params map[string]string, options ...Option) (*jnode.Node, error)
	Delete(path string, options ...Option) (*jnode.Node, error)
	XCPPost(orgID string, module string, files []string, values map[string]string, options ...Option) error
	GetClient() *resty.Client
	GetOrganization() string
}

type clientT struct {
	*resty.Client
	Config
}

func NewClient(config *Config) Interface {
	c := &clientT{
		Client: resty.New(),
		Config: *config,
	}
	c.Token = config.APIToken
	if c.APIPrefix == "" {
		c.APIPrefix = "/api/v1"
	}

	apiServer := config.APIServer
	c.SetHostURL(apiServer)
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
	c.OnBeforeRequest(func(rc *resty.Client, r *resty.Request) error {
		if strings.Contains(r.URL, orgToken) {
			if c.Organization == "" {
				log.Errorf("An organization must be specified with --organization or configuring one with `cli-config set organization <org-id>`")
				return fmt.Errorf("organization is required")
			}
			r.URL = strings.ReplaceAll(r.URL, orgToken, c.Organization)
		}
		if len(r.URL) > 0 && r.URL[0] != '/' {
			r.URL = fmt.Sprintf("%s/%s", c.APIPrefix, r.URL)
		}
		log.Debugf("{primary:%s %s}", r.Method, r.URL)
		var b strings.Builder
		_ = r.Header.Write(&b)
		return nil
	})

	c.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		info := r.Request.TraceInfo()
		log.Debugf("{warning:%+v}\n", info)
		return nil
	})

	c.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		if r.IsError() {
			log.Errorf("{info:%s} {primary:%s} returned {danger:%d}\n", r.Request.Method,
				r.Request.URL, r.StatusCode())
			log.Errorf("{warning:%s}\n", r.String())
			return fmt.Errorf("%s returned %d", r.Request.URL, r.StatusCode())
		}
		log.Debugf("%v", r.Result())
		log.Infof("{info:%s} {primary:%s} returned {success:%d}\n", r.Request.Method,
			r.Request.URL, r.StatusCode())
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

func applyOptions(r *resty.Request, options []Option) *resty.Request {
	for _, opt := range options {
		opt(r)
	}
	return r
}

func (c *clientT) Post(path string, body *jnode.Node, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()
	if _, err := applyOptions(c.R().SetBody(body).SetResult(result), options).Post(path); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *clientT) Get(path string, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()
	if _, err := applyOptions(c.R().SetResult(result), options).Get(path); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *clientT) GetWithParams(path string, params map[string]string, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()
	if _, err := applyOptions(c.R().SetQueryParams(params).SetResult(result), options).Get(path); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *clientT) Delete(path string, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()
	if _, err := applyOptions(c.R().SetResult(result), options).Delete(path); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *clientT) Patch(path string, body *jnode.Node, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()
	if _, err := applyOptions(c.R().SetResult(result).SetBody(body), options).Patch(path); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *clientT) GetClient() *resty.Client {
	return c.Client
}

func (c *clientT) XCPPost(orgID string, module string, files []string, values map[string]string, options ...Option) error {
	if module == "" {
		return fmt.Errorf("module parameter is required")
	}
	req := c.R()
	for i, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		req.SetFileReader(fmt.Sprintf("file_%d", i), filepath.Base(file), f)
	}
	req.SetHeader("X-SOLUBLE-ORG-ID", orgID)
	req.SetMultipartFormData(values)
	req = applyOptions(req, options)
	_, err := req.Post(fmt.Sprintf("/api/v1/xcp/%s/data", module))
	return err
}

func (c *clientT) GetOrganization() string {
	return c.Organization
}
