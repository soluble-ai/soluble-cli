package download

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/soluble-ai/soluble-cli/pkg/archive"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type Manager struct {
	meta        []*DownloadMeta
	downloadDir string
}

type Download struct {
	Name        string
	Version     string
	URL         string
	Dir         string
	InstallTime time.Time
}

type DownloadMeta struct {
	Name            string
	LatestVersion   string
	LatestCheckTime time.Time
	Installed       []*Download
	Dir             string
}

func NewManager() *Manager {
	return &Manager{
		downloadDir: filepath.Join(config.ConfigDir, "downloads"),
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
				data, err := ioutil.ReadFile(path)
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

func (m *Manager) InstallGithubRelease(owner, repo, tag string) (*Download, error) {
	name := fmt.Sprintf("%s-%s", owner, repo)
	meta := m.findOrCreateMeta(name)
	v := meta.FindVersion(tag)
	if v != nil {
		return v, nil
	}
	release, asset, err := getGithubReleaseAsset(owner, repo, tag)
	if err != nil {
		return nil, err
	}
	if latest := meta.updateLatestInfo(tag, release.GetTagName()); latest != nil {
		return latest, nil
	}
	return meta.install(m, release.GetTagName(), asset.GetBrowserDownloadURL())
}

func (m *Manager) Install(name, version, url string, options ...DownloadOption) (*Download, error) {
	meta := m.findOrCreateMeta(name)
	v := meta.FindVersion(version)
	if v != nil {
		return v, nil
	}
	return meta.install(m, version, url, options...)
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

func (meta *DownloadMeta) updateLatestInfo(tag, version string) *Download {
	if isLatestTag(tag) {
		meta.LatestCheckTime = time.Now()
		meta.LatestVersion = version
		log.Infof("Latest release of {secondary:%s} is {info:%s}", meta.Name, meta.LatestVersion)
		return meta.findVersionExactly(version)
	}
	return nil
}

func (meta *DownloadMeta) install(m *Manager, version, url string, options ...DownloadOption) (*Download, error) {
	base, err := getBaseName(url)
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
	log.Infof("Getting {info:%s}", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for _, opt := range options {
		if err = opt(req); err != nil {
			return nil, err
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned %d", url, resp.StatusCode)
	}
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return nil, err
	}
	d := &Download{
		Name:        meta.Name,
		Version:     version,
		URL:         url,
		Dir:         filepath.Join(m.downloadDir, meta.Name, version),
		InstallTime: time.Now(),
	}
	meta.removeInstalledVersion(d.Version)
	meta.Installed = append(meta.Installed, d)
	err = d.Install(archiveFile)
	if err != nil {
		return nil, err
	}
	err = m.save(meta)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (meta *DownloadMeta) FindVersion(version string) *Download {
	if isLatestTag(version) {
		if meta.LatestVersion != "" && meta.LatestCheckTime.After(time.Now().Add(-24*time.Hour)) {
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
	v := meta.FindVersion(version)
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

func (d *Download) Install(file string) error {
	base := filepath.Base(file)
	var unpack archive.Unpack
	switch {
	case strings.HasSuffix(base, ".tar.gz"):
		fallthrough
	case strings.HasSuffix(base, ".tar"):
		unpack = archive.Untar
	case strings.HasSuffix(base, ".zip"):
		unpack = archive.Unzip
	default:
		return fmt.Errorf("unknown archive format %s", base)
	}
	log.Infof("Installing {info:%s}", base)
	return archive.Do(unpack, file, d.Dir, nil)
}

func getBaseName(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	return filepath.Base(u.Path), nil
}
