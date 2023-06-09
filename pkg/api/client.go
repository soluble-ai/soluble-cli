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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/lacework/go-sdk/api"
	"github.com/soluble-ai/go-jnode"
	cfg "github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/util"
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
	NoOrganizationHook func() error
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

var ErrNoContent = errors.New("no content")

func NewClient(config *Config) *Client {
	c := &Client{
		Client: resty.New(),
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

func (c *Client) configureAuthHeaders(headers http.Header) error {
	if c.LaceworkAccount != "" {
		if err := c.ensureLaceworkAPIToken(); err != nil {
			return err
		}
		headers.Set("X-LW-Domain", c.GetDomain())
		headers.Set("X-LW-Authorization", fmt.Sprintf("Token %s", c.LaceworkAPIToken))
		log.Debugf("Using lacework authentication")
	} else if c.LegacyAPIToken != "" {
		headers.Set("Authorization", fmt.Sprintf("Bearer %s", c.LegacyAPIToken))
	}
	return nil
}

func (c *Client) execute(r *resty.Request, method, path string, options []Option) (*resty.Response, error) {
	// set r.Method here so that options can do different things
	// depending on the method
	r.Method = method
	for _, opt := range options {
		opt.Apply(r)
	}
	if strings.Contains(path, orgToken) || strings.HasPrefix(path, "/api/v1/xcp/") {
		if c.Organization == "" && c.NoOrganizationHook != nil {
			if err := c.NoOrganizationHook(); err != nil {
				return nil, err
			}
		}
		if c.Organization == "" {
			return nil, fmt.Errorf("an IAC organization is required")
		}
	}
	if strings.Contains(path, orgToken) {
		path = strings.ReplaceAll(path, orgToken, c.Organization)
	}
	if c.Organization != "" {
		r.SetHeader("X-SOLUBLE-ORG-ID", c.Organization)
	}
	if err := c.configureAuthHeaders(r.Header); err != nil {
		return nil, err
	}
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
		return nil, err
	case resp != nil && resp.IsError():
		t := time.Since(r.Time).Truncate(time.Millisecond)
		if shouldLog(path) {
			log.Errorf("{info:%s} {primary:%s} returned {danger:%d} in {secondary:%s}\n", r.Method,
				r.URL, resp.StatusCode(), t)
			log.Errorf("{warning:%s}\n", resp.String())
			if resp.StatusCode() == 401 || resp.StatusCode() == 404 {
				log.Infof("Are you not logged in?  Use {info:%s auth profile} to verify.", cfg.CommandInvocation())
				log.Infof("See {primary:https://docs.lacework.com/iac/} for more information.")
			}
		}
		return resp, httpError(fmt.Sprintf("%s returned %d", r.URL, resp.StatusCode()))
	default:
		t := time.Since(r.Time).Truncate(time.Millisecond)
		log.Tracef("%v", resp.Result())
		if shouldLog(path) {
			log.Infof("{info:%s} {primary:%s} returned {success:%d} in {secondary:%s}\n", r.Method,
				r.URL, resp.StatusCode(), t)
		}
		return resp, nil
	}
}

func shouldLog(path string) bool {
	if strings.Contains(path, "cli/tools/") && strings.HasSuffix(path, "/config") {
		return false
	}
	return true
}

func (c *Client) Post(path string, body *jnode.Node, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()
	if _, err := c.execute(c.R().SetBody(body).SetResult(result), resty.MethodPost, path, options); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) Get(path string, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()
	if _, err := c.execute(c.R().SetResult(result), resty.MethodGet, path, options); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetWithParams(path string, params map[string]string, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()

	if _, err := c.execute(c.R().SetQueryParams(params).SetResult(result), resty.MethodGet, path, options); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) Delete(path string, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()
	if _, err := c.execute(c.R().SetResult(result), resty.MethodDelete, path, options); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) Patch(path string, body *jnode.Node, options ...Option) (*jnode.Node, error) {
	result := jnode.NewObjectNode()
	if _, err := c.execute(c.R().SetResult(result).SetBody(body), resty.MethodPatch, path, options); err != nil {
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
	if _, err := c.execute(req, resty.MethodPost, fmt.Sprintf("/api/v1/xcp/%s/data", module), options); err != nil {
		return nil, err
	}
	// also post results to cds using the cdk, if it is configured as lacework component
	if cfg.IsRunningAsComponent() {
		// if files are not present directly then look in request and get the files to upload
		// most of the tools are adding the multipart files in the options so extract them from the request and send it to CDS
		files, _ := getFilesForCDS(req, files, values, result)
		_ = uploadResultsToCDS(c, files)
		// if err != nil {
		// log.Errorf("upload failed %s", err)
		// CDS upload shouldn't block the other things at the moment
		// return nil, err
		// }
	} else {
		log.Debugf("Skipping the upload of results to CDS")
	}
	return result, nil
}

// function to upload results to CDS, if the iac is configured as component under lacework cli
func uploadResultsToCDS(c *Client, filesToUpload []string) error {
	lwAPI, err := api.NewClient(c.Config.LaceworkAccount,
		api.WithApiKeys(c.Config.LaceworkAPIKey, c.Config.LaceworkAPISecret),
		api.WithApiV2(),
	)
	if err != nil {
		return err
	}
	log.Infof("Uploading %d files to CDS", len(filesToUpload))
	if len(filesToUpload) > 0 {
		guid, err := lwAPI.V2.ComponentData.UploadFiles("iac-results", []string{"iac"}, filesToUpload)
		if err != nil {
			// log.Errorf("{warning:Unable to upload results %s\n}", err)
			return err
		}
		log.Infof("Successfully uploaded to CDS: {info:%s}", guid)
	}
	return nil
}

func (c *Client) Download(path string) ([]byte, error) {
	req := c.R()
	resp, err := c.execute(req, resty.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusNoContent {
		return nil, ErrNoContent
	}
	return resp.Body(), nil
}

func (c *Client) GetOrganization() string {
	return c.Organization
}

func (c *Client) GetHostURL() string {
	return c.APIServer
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

// function to get the dirty work done for writing files to CDS
func getFilesForCDS(req *resty.Request, files []string, values map[string]string, result *jnode.Node) ([]string, error) {
	if len(files) == 0 {
		for _, v := range req.FormData {
			files = append(files, v[0])
		}

		// add the metadata/env variables from the values as Json object to metadata.json file and upload that as well
		// this is missing with CDS as it doesn't upload any environment variables
		// convert the map to a JSON encoded byte slice
		jsonContent, err := json.Marshal(values)
		if err != nil {
			return nil, err
		}
		envVariablesFile, err := writeToFile("env_variables.json", jsonContent)
		if err != nil {
			return nil, err
		}
		files = append(files, envVariablesFile)

    assessment := result.Path("assessment")

		if assessment != jnode.MissingNode {
			// add the enhanced result json file also to the CDS upload
			enrichedResultsFile, err := writeToFile("enriched_results.json", []byte(assessment.String()))
			if err != nil {
				return nil, err
			}
			files = append(files, enrichedResultsFile)
		}
	}
	return files, nil
}

func writeToFile(filename string, content []byte) (string, error) {
	path, _ := util.GetTempFilePath(filename)
	if err := os.WriteFile(path, content, 0600); err != nil {
		return "", err
	}
	return path, nil
}
