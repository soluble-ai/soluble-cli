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

package xcp

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/compress"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/util"

	"github.com/stretchr/testify/assert"
)

const (
	readerTestData = "Hello Laceworld!"
)

func setupTest(t *testing.T) func(t *testing.T) {
	log.Infof(fmt.Sprintf("setup %s", t.Name()))
	_ = util.CreateRootTempDir()

	return func(t *testing.T) {
		log.Infof(fmt.Sprintf("teardown %s", t.Name()))
		util.RemoveRootTempDir()
	}
}

func TestGetCIEnv(t *testing.T) {
	assertions := assert.New(t)
	saveEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, kv := range saveEnv {
			eq := strings.Index(kv, "=")
			os.Setenv(kv[0:eq], kv[eq+1:])
		}
	}()
	// xxx must not be included, yyy must be included
	os.Setenv("PASSWORD", "xxx")
	os.Setenv("GITHUB_TOKEN", "xxx")
	os.Setenv("GITHUB_BRANCH", "yyy")
	os.Setenv("BUILDKITE_AGENT_ACCESS_TOKEN", "xxx")
	os.Setenv("BUILDKITE_COMMAND", "xxx")
	os.Setenv("BUILDKITE_S3_ACCESS_URL", "xxx")
	os.Setenv("BITBUCKET_BUILD_NUMBER", "yyy")
	os.Setenv("BITBUCKET_STEP_OIDC_TOKEN", "xxx")
	os.Setenv("ATLANTIS_TERRAFORM_VERSION", "yyy")
	os.Setenv("PULL_NUM", "yyy")
	os.Setenv("REPO_REL_DIR", "yyy")
	os.Setenv("BUILD_ID", "27")
	os.Setenv("JOB_BASE_NAME", "main")
	os.Setenv("KUBERNETES_PORT", "tcp://172.20.0.1:443")
	os.Setenv("RUN_ARTIFACTS_DISPLAY_URL", "https://ci.intouchhealth.io/")
	os.Setenv("TF_VAR_adminpassword", "****")
	os.Setenv("TF_VAR_adminusername", "***")
	os.Setenv("TF_VAR_adminusername", "***")
	os.Setenv("ARM_TENANT_ID", "****")
	os.Setenv("ARM_CLIENT_SECRET", "yyy")
	env := GetCIEnv(".")
	for k, v := range env {
		if v == "xxx" {
			t.Error(k, v)
		}
	}

	assertions.True(contains(env, "BUILD_ID"))
	assertions.True(contains(env, "JOB_BASE_NAME"))
	assertions.True(contains(env, "KUBERNETES_PORT"))
	assertions.True(contains(env, "SOLUBLE_METADATA_CI_SYSTEM"))
	assertions.True(contains(env, "RUN_ARTIFACTS_DISPLAY_URL"))
	assertions.False(contains(env, "TF_VAR_adminpassword"))
	assertions.False(contains(env, "TF_VAR_adminusername"))
	assertions.False(contains(env, "ARM_TENANT_ID"))
	assertions.False(contains(env, "ARM_CLIENT_SECRET"))
}

func TestAtlantisCIEnv(t *testing.T) {
	assertions := assert.New(t)
	saveEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, kv := range saveEnv {
			eq := strings.Index(kv, "=")
			os.Setenv(kv[0:eq], kv[eq+1:])
		}
	}()

	os.Setenv("ATLANTIS_TERRAFORM_VERSION", "yyy")
	os.Setenv("PULL_NUM", "yyy")
	os.Setenv("REPO_REL_DIR", "yyy")
	os.Setenv("BUILD_ID", "27")
	os.Setenv("JOB_BASE_NAME", "main")
	env := GetCIEnv(".")
	for k, v := range env {
		if v == "xxx" {
			t.Error(k, v)
		}
	}

	// make sure atlantis env variables are available
	assertions.True(contains(env, "ATLANTIS_TERRAFORM_VERSION"))
	assertions.True(contains(env, "ATLANTIS_PULL_NUM"))
	for _, kv := range os.Environ() {
		if strings.HasSuffix(kv, "=yyy") {
			if strings.HasPrefix(kv, "PULL_NUM") ||
				strings.HasPrefix(kv, "REPO_REL_DIR") {
				kv = "ATLANTIS_" + kv
			}
			if env[kv[0:len(kv)-4]] != "yyy" {
				t.Error(kv)
			}
		}
	}
}

func TestNormalizeGitRemote(t *testing.T) {
	if s := normalizeGitRemote("git@github.com:fizz/buzz.git"); s != "github.com/fizz/buzz" {
		t.Error(s)
	}
}

func TestCompressedGetFileReaderFromStringReader(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)
	reader := strings.NewReader(readerTestData)
	gzReader := compress.NewGZIPPipe(reader)
	// assert we can read from the call to writeFileFromReader(filename string, reader io.Reader)
	assertWriteCompressedFileFromReader(t, "compressedfile", gzReader)
}

func TestGetFileReaderFromStringReader(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)
	reader := strings.NewReader(readerTestData)
	// assert we can read from the call to writeFileFromReader(filename string, reader io.Reader)
	assertWriteFileFromReader(t, "afile", reader)
}

func TestCompressedGetFileReaderFromFile(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)
	file, err := os.CreateTemp("", "helloworld.txt")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	err = os.WriteFile(file.Name(), []byte(readerTestData), 0600)
	assert.NoError(t, err)
	gzReader := compress.NewGZIPPipe(file)
	// assert we can read from the call to writeFileFromReader(filename string, reader io.Reader)
	assertWriteCompressedFileFromReader(t, "compressedfile", gzReader)
}

func TestGetFileReaderFromFile(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)
	file, err := os.CreateTemp("", "helloworld.txt")
	assert.NoError(t, err)
	defer os.Remove(file.Name())
	err = os.WriteFile(file.Name(), []byte(readerTestData), 0600)
	assert.NoError(t, err)
	// assert we can read from the call to writeFileFromReader(filename string, reader io.Reader)
	assertWriteFileFromReader(t, "afile", file)
}

func TestCompressedGetFileReaderFromBytesReader(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)
	reader := bytes.NewReader([]byte(readerTestData))
	gzReader := compress.NewGZIPPipe(reader)
	// assert we can read from the call to writeFileFromReader(filename string, reader io.Reader)
	assertWriteCompressedFileFromReader(t, "compressedfile", gzReader)
}

func TestGetFileReaderFromBytesReader(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)
	reader := bytes.NewReader([]byte(readerTestData))
	// assert we can read from the call to writeFileFromReader(filename string, reader io.Reader)
	assertWriteFileFromReader(t, "afile", reader)
}

func assertWriteCompressedFileFromReader(t *testing.T, filename string, reader io.Reader) {
	// should always return a reader which can be read
	path, err := writeFileFromReader(filename, reader)
	assert.NoError(t, err)
	file, err := os.Open(path)
	assert.NoError(t, err)
	// gzip reader to read the compressed data
	gzr, err := gzip.NewReader(file)
	assert.NoError(t, err)
	// read all the data
	bytesRead, err := ioutil.ReadAll(gzr)
	assert.NoError(t, err)
	// convert to string
	actualData := string(bytesRead)
	// assert that the actual data read meets the expected data from the io.readSeeker
	assert.Equal(t, readerTestData, actualData)
}

func assertWriteFileFromReader(t *testing.T, filename string, reader io.Reader) {
	// should always return a reader which can be read
	path, err := writeFileFromReader(filename, reader)
	assert.NoError(t, err)
	file, err := os.Open(path)
	assert.NoError(t, err)
	// read all the data
	bytesRead, err := ioutil.ReadAll(file)
	assert.NoError(t, err)
	// convert to string
	actualData := string(bytesRead)
	// assert that the actual data read meets the expected data from the io.readSeeker
	assert.Equal(t, readerTestData, actualData)
}

func contains(s map[string]string, searchStr string) bool {
	for k := range s {
		if k == searchStr {
			return true
		}
	}
	return false
}
