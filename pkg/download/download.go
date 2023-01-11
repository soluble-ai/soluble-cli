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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/soluble-ai/soluble-cli/pkg/archive"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/download/gcs"
	"github.com/soluble-ai/soluble-cli/pkg/download/terraform"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/repotree"
	"github.com/spf13/afero"
)

type Manager struct {
	meta        []*DownloadMeta
	downloadDir string
}

type Download struct {
	Name              string
	Version           string
	URL               string
	APIServerArtifact string
	Dir               string
	InstallTime       time.Time
	OverrideExe       string `json:"-"`
}

type DownloadMeta struct {
	Name            string
	LatestVersion   string
	LatestCheckTime time.Time
	Installed       []*Download
	Dir             string
}

type Spec struct {
	Name                       string
	RequestedVersion           string
	URL                        string
	APIServerArtifact          string
	APIServer                  APIServer
	GithubReleaseMatcher       GithubReleaseMatcher
	LatestReleaseCacheDuration time.Duration
	GetLatestVersion           func(*Spec) (string, error)
}

type APIServer interface {
	GetHostURL() string
	Download(path string) ([]byte, error)
}

type urlResolverFunc func(requestedVersion string) (version string, url string, err error)

var urlResolvers = map[string]urlResolverFunc{
	"terraform": terraform.GetVersionAndURL,
	"tfscore":   gcs.NewResolver("soluble-public", "tfscore"),
	"opal":      gcs.NewResolver("soluble-public", "opal"),
}

func NewManager() *Manager {
	var downloadDir string
	if config.ConfigDir == "" {
		// we are presumably running within a test, so we'll use
		// <repo-root>/.downloads
		root, err := repotree.FindRepoRoot("")
		if err != nil {
			panic(err)
		}
		downloadDir = filepath.Join(root, ".downloads")
	} else {
		downloadDir = filepath.Join(config.ConfigDir, "downloads")
	}
	return &Manager{
		downloadDir: downloadDir,
	}
}

func (m *Manager) GetMeta(name string) *DownloadMeta {
	for _, meta := range m.List() {
		if meta.Name == name {
			return meta
		}
	}
	return nil
}

func (m *Manager) List() (result []*DownloadMeta) {
	if m.meta == nil {
		_ = filepath.Walk(m.downloadDir, func(path string, info os.FileInfo, err1 error) error {
			if info == nil {
				return nil
			}
			r, err := filepath.Rel(m.downloadDir, path)
			if err != nil {
				return err
			}
			if info.IsDir() {
				if strings.Count(r, string(filepath.Separator)) > 0 {
					return filepath.SkipDir
				}
				return nil
			}
			_, file := filepath.Split(r)
			if file == "meta.json" {
				var m DownloadMeta
				data, err := os.ReadFile(path)
				if err == nil {
					if json.Unmarshal(data, &m) == nil {
						result = append(result, &m)
					}
				}
			}
			return nil
		})
		m.meta = result
	}
	result = m.meta
	return
}

func (m *Manager) findOrCreateMeta(name string) *DownloadMeta {
	meta := m.GetMeta(name)
	if meta != nil {
		return meta
	}
	return &DownloadMeta{
		Name: name,
		Dir:  filepath.Join(m.downloadDir, name),
	}
}

func (m *Manager) Reinstall(spec *Spec) (*Download, error) {
	name := spec.Name
	if name == "" {
		owner, repo := parseGithubRepo(spec.URL)
		if owner != "" {
			name = fmt.Sprintf("%s-%s", owner, repo)
		}
	}
	meta := m.GetMeta(name)
	if meta != nil {
		d := meta.FindVersion(spec.RequestedVersion, 0, true)
		_ = m.Remove(name, spec.RequestedVersion)
		if d != nil {
			if spec.URL == "" {
				spec.URL = d.URL
			}
			if spec.APIServerArtifact == "" {
				spec.APIServerArtifact = d.APIServerArtifact
			}
		}
	}
	return m.Install(spec)
}

func (m *Manager) Install(spec *Spec) (*Download, error) {
	owner, repo := parseGithubRepo(spec.URL)
	if owner != "" {
		spec.Name = fmt.Sprintf("%s-%s", owner, repo)
	}
	if spec.Name == "" {
		return nil, fmt.Errorf("name must be specified for plain URL downloads")
	}
	// see if we've already installed it
	meta := m.findOrCreateMeta(spec.Name)
	v := meta.FindVersion(spec.RequestedVersion, spec.LatestReleaseCacheDuration, false)
	if v != nil {
		return v, nil
	}
	actualVersion := spec.RequestedVersion
	if urf := urlResolvers[spec.Name]; urf != nil {
		var err error
		actualVersion, spec.URL, err = urf(spec.RequestedVersion)
		if err != nil {
			return nil, err
		}
	}
	if owner != "" {
		// find the github release
		release, asset, err := getGithubReleaseAsset(owner, repo, spec.RequestedVersion, spec.GithubReleaseMatcher)
		if err != nil {
			return nil, err
		}
		actualVersion = release.GetTagName()
		if owner == "helm" && repo == "helm" {
			spec.URL = getHelmDownloadURL(asset)
		} else {
			spec.URL = asset.GetBrowserDownloadURL()
		}
		if latest := meta.updateLatestInfo(spec.RequestedVersion, actualVersion); latest != nil {
			// if we've requested "latest" and we've already got that specific version
			// installed, then just update the latest check time and we're done
			_ = m.save(meta)
			return latest, nil
		}
	}
	options := []downloadOption{}
	if spec.APIServerArtifact != "" {
		spec.URL = fmt.Sprintf("%s/%s", spec.APIServer.GetHostURL(), spec.APIServerArtifact)
		actualVersion = "latest"
	}
	if spec.URL == "" {
		return nil, fmt.Errorf("download URL must be specified")
	}
	return meta.install(m, spec, actualVersion, options)
}

func (m *Manager) Remove(name, version string) error {
	meta := m.GetMeta(name)
	if meta == nil {
		return nil
	}
	if version == "" {
		log.Infof("Removing {info:%s}", meta.Dir)
		if err := os.RemoveAll(meta.Dir); err != nil {
			return err
		}
		m.meta = nil
		return nil
	}
	return meta.removeVersion(m, version)
}

func (m *Manager) save(meta *DownloadMeta) error {
	f, err := os.Create(filepath.Join(m.downloadDir, meta.Name, "meta.json"))
	if err != nil {
		return err
	}
	defer f.Close()
	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	err = e.Encode(meta)
	if err != nil {
		return err
	}
	m.meta = nil
	return nil
}

func (meta *DownloadMeta) updateLatestInfo(requestedVersion, actualVersion string) *Download {
	if isLatestTag(requestedVersion) {
		meta.LatestCheckTime = time.Now()
		meta.LatestVersion = actualVersion
		log.Debugf("Latest release of {primary:%s} is {info:%s}", meta.Name, meta.LatestVersion)
		return meta.findVersionExactly(actualVersion)
	}
	return nil
}

func download(spec *Spec, req *http.Request, name string) (io.Reader, error) {
	if spec.APIServerArtifact != "" {
		resp, err := spec.APIServer.Download(spec.APIServerArtifact)
		if err != nil {
			return nil, err
		}
		body := bytes.NewReader(resp)
		return body, nil
	} else {
		req.BasicAuth()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			log.Errorf("Request to install {warning:%s} returned status code {danger:%d}", name,
				resp.StatusCode)
			return nil, fmt.Errorf("%s returned %d", spec.URL, resp.StatusCode)
		}
		return resp.Body, nil
	}
}
func (meta *DownloadMeta) install(m *Manager, spec *Spec, actualVersion string, options []downloadOption) (*Download, error) {
	base, err := getBaseName(spec.URL)
	if err != nil {
		return nil, err
	}
	nameDir := filepath.Join(m.downloadDir, meta.Name)
	if err := os.MkdirAll(nameDir, 0777); err != nil {
		return nil, err
	}
	archiveFile := filepath.Join(nameDir, base)
	w, err := os.Create(archiveFile)
	if err != nil {
		return nil, err
	}
	defer w.Close()
	log.Infof("Getting {info:%s}", spec.URL)
	req, err := http.NewRequest("GET", spec.URL, nil)
	if err != nil {
		return nil, err
	}
	for _, opt := range options {
		if err = opt(req); err != nil {
			return nil, err
		}
	}
	body, err := download(spec, req, meta.Name)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(w, body)
	if err != nil {
		return nil, err
	}
	d := &Download{
		Name:              meta.Name,
		Version:           actualVersion,
		URL:               spec.URL,
		APIServerArtifact: spec.APIServerArtifact,
		// remove special fs characters from tag
		Dir:         filepath.Join(m.downloadDir, meta.Name, noslashdotdots(actualVersion)),
		InstallTime: time.Now(),
	}
	meta.removeInstalledVersion(d.Version)
	meta.Installed = append(meta.Installed, d)
	err = d.Install(archiveFile)
	if err != nil {
		return nil, err
	}
	meta.updateLatestInfo(spec.RequestedVersion, actualVersion)
	err = m.save(meta)
	if err != nil {
		return nil, err
	}

	return d, nil
}

var slashDotDotsRe = regexp.MustCompile(`(\.\.)|[/\\]`)

func noslashdotdots(s string) string {
	return slashDotDotsRe.ReplaceAllString(s, "_")
}

func (meta *DownloadMeta) FindVersion(version string, cacheTime time.Duration, stale bool) *Download {
	if isLatestTag(version) {
		if cacheTime == 0 {
			cacheTime = 24 * time.Hour
		}
		if meta.LatestVersion != "" && (stale || meta.LatestCheckTime.After(time.Now().Add(-cacheTime))) {
			version = meta.LatestVersion
		} else {
			return nil
		}
	}
	return meta.findVersionExactly(version)
}

func (meta *DownloadMeta) findVersionExactly(version string) *Download {
	for _, v := range meta.Installed {
		if v.Version == version {
			// check to see that it's still there
			_, err := os.Stat(v.Dir)
			if err != nil {
				log.Warnf("{warning:%s} is no longer accesible", v.Dir)
				return nil
			}
			return v
		}
	}
	return nil
}

func (meta *DownloadMeta) removeVersion(m *Manager, version string) error {
	v := meta.FindVersion(version, 0, false)
	if v != nil {
		log.Infof("Removing {info:%s}", v.Dir)
		err := os.RemoveAll(v.Dir)
		if err != nil {
			return err
		}
		meta.removeInstalledVersion(v.Version)
		return m.save(meta)
	}
	return nil
}

func (meta *DownloadMeta) removeInstalledVersion(version string) {
	var installed []*Download
	for _, iv := range meta.Installed {
		if iv.Version != version {
			installed = append(installed, iv)
		}
	}
	meta.Installed = installed
}

func (meta *DownloadMeta) FindLatestOrLastInstalledVersion() *Download {
	if meta.LatestVersion != "" {
		return meta.findVersionExactly(meta.LatestVersion)
	}
	var v *Download
	for _, d := range meta.Installed {
		if v == nil || d.InstallTime.After(v.InstallTime) {
			v = d
		}
	}
	return v
}

func (d *Download) Install(file string) error {
	base := filepath.Base(file)
	var unpack archive.Unpack
	switch {
	case strings.HasSuffix(base, ".tgz"):
		fallthrough
	case strings.HasSuffix(base, ".tar.gz"):
		fallthrough
	case strings.HasSuffix(base, ".tar"):
		unpack = archive.Untar
	case strings.HasSuffix(base, ".zip"):
		unpack = archive.Unzip
	case !strings.Contains(base, ".") || strings.HasSuffix(base, ".exe"):
		unpack = d.installExecutable
	default:
		return fmt.Errorf("unknown archive format %s", base)
	}
	log.Infof("Installing {info:%s}", base)
	return archive.Do(unpack, file, d.Dir, nil)
}

func (d *Download) GetExePath(path string) string {
	if d.OverrideExe != "" {
		return d.OverrideExe
	}
	exe := filepath.Join(d.Dir, path)
	if _, err := os.Stat(exe); err != nil && errors.Is(err, os.ErrNotExist) {
		// if the unarchived dir contains a single dir, then look for the
		// the path in dir e.g. archive puts its content in a subdir
		entries, err := os.ReadDir(d.Dir)
		if err != nil {
			log.Warnf("Could not read {warning:%s}", d.Dir)
		} else {
			for _, entry := range entries {
				if entry.IsDir() {
					direxe := filepath.Join(d.Dir, entry.Name(), path)
					if _, err := os.Stat(direxe); err == nil {
						return direxe
					}
				}
			}
		}
	}
	return exe
}

func (d *Download) installExecutable(src afero.File, fs afero.Fs, options *archive.Options) error {
	// just copy the file
	name := d.Name
	if owner, repo := parseGithubRepo(d.URL); owner != "" {
		name = repo
	}
	if strings.HasSuffix(src.Name(), ".exe") {
		name = fmt.Sprintf("%s.exe", name)
	}
	out, err := fs.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

func getBaseName(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	return filepath.Base(u.Path), nil
}
