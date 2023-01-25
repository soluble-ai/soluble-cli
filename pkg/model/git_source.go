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

package model

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type GitSource struct {
	FileSystemSource
	WasFetched bool
	version    string
}

func GetGitSource(url string) (Source, error) {
	dir, err := getGitModelDir(url)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0700); err != nil && !os.IsExist(err) {
		return nil, err
	}
	gitConfig, err := os.Stat(filepath.Join(dir, ".git", "config"))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	var fetchHead os.FileInfo
	if gitConfig != nil {
		fetchHead, _ = os.Stat(filepath.Join(dir, ".git", "FETCH_HEAD"))
	}
	var wasFetched bool
	switch {
	case gitConfig == nil:
		// repo doesn't exist, clone the repo
		log.Debugf("Cloning {primary:%s} to {info:%s}", url, dir)
		err := git("clone", "--depth", "1", url, dir).Run()
		if err != nil {
			return nil, err
		}
	case fetchHead == nil || time.Now().After(fetchHead.ModTime().Add(5*time.Minute)):
		// repo exists, and we haven't fetched it in a while
		log.Debugf("Updating git model repository {primary:%s}", dir)
		c := git("fetch", "-q", "--depth", "1")
		c.Dir = dir
		done := make(chan error)
		go run(c, done)
		select {
		case err = <-done:
			if err != nil {
				log.Warnf("Could not fetch {primary:%s}: {warning:%s}", url, err.Error())
			} else {
				wasFetched = true
			}
		case <-time.After(15 * time.Second):
			log.Warnf("Fetching {primary:%s} is taking a while, killing the fetch", url)
			_ = c.Process.Kill()
		}
	}

	source := &GitSource{
		FileSystemSource: FileSystemSource{
			Filesystem: os.DirFS(dir),
			RootPath:   url,
		},
		WasFetched: wasFetched,
		version:    headVersion(dir),
	}
	return source, nil
}

func (s *GitSource) GetVersion(name string, content []byte) string {
	return s.version
}

func getGitModelDir(url string) (string, error) {
	hash := sha256.Sum256([]byte(url))
	name := fmt.Sprintf("%012x", hash[0:6])
	m := filepath.Join(config.ConfigDir, "iac", "cli-models")
	return filepath.Join(m, name), nil
}

func headVersion(dir string) string {
	c := exec.Command("git", "describe", "--always", "--dirty")
	c.Dir = dir
	out, err := c.Output()
	if err != nil {
		return "?"
	}
	return strings.TrimRight(string(out), "\n\r")
}

func git(args ...string) *exec.Cmd {
	c := exec.Command("git", args...)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	return c
}

func run(c *exec.Cmd, done chan error) {
	done <- c.Run()
}
