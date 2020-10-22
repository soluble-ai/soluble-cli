package download

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestDownload(t *testing.T) {
	setupHTTP()
	m := setupManager()
	if result := m.List(); len(result) != 0 {
		t.Error("non-empty list")
	}
	d, err := m.Install("hello", "1.0", "https://example.com/hello.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	result := m.List()
	if len(result) != 1 {
		t.Error("after install, wrong # of results", result)
	}
	meta := m.GetMeta(d.Name)
	if meta == nil {
		t.Error("no meta after install")
	}
	_, err = m.Install("hello", "2.0", "https://example.com/hello.tar.gz")
	if err != nil {
		t.Error(err)
	}
	meta = m.GetMeta(d.Name)
	if meta == nil || len(meta.Installed) != 2 {
		t.Error("expecting 2 versions")
	}
	if err = m.Remove("hello", "1.0"); err != nil {
		t.Error(err)
	}
	meta = m.GetMeta(d.Name)
	if meta == nil || len(meta.Installed) != 1 || meta.Installed[0].Version != "2.0" {
		t.Error("expecting only version 2 now")
	}
	if err = m.Remove("hello", ""); err != nil {
		t.Error(err)
	}
	meta = m.GetMeta(d.Name)
	if meta != nil {
		t.Error("expecting hello removed", meta)
	}
}

func TestDownloadZip(t *testing.T) {
	setupHTTP()
	m := setupManager()
	d, err := m.Install("hello-zip", "1.0", "https://example.com/hello.zip")
	if err != nil {
		t.Error(err)
	}
	_, err = os.Stat(filepath.Join(d.Dir, "README.txt"))
	if err != nil {
		t.Error(err)
	}
}

func setupManager() *Manager {
	m := NewManager()
	dir, err := ioutil.TempDir("", "downloadtest*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	m.downloadDir = dir
	return m
}

func setupHTTP() {
	httpmock.Activate()
	registerTestArchive("hello.tar.gz")
	registerTestArchive("hello.zip")
}

func registerTestArchive(name string) {
	dat, err := ioutil.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		panic(err)
	}
	httpmock.RegisterResponder("GET", fmt.Sprintf("https://example.com/%s", name),
		httpmock.NewBytesResponder(200, dat))
}
