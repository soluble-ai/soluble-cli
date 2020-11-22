package tools

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/archive"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/afero"
)

func TestFilesTarball(t *testing.T) {
	dir, _ := ioutil.TempDir("", "toolopt*")
	defer os.RemoveAll(dir)
	createFile(dir, "zero.txt", "zero\n")
	createFile(dir, "one/one.txt", "one\n")
	result := &Result{
		Directory: dir,
	}
	result.AddFile("zero.txt")
	result.AddFile(filepath.Join(dir, "one/one.txt"))
	f, err := result.createTarball()
	util.Must(err)
	mfs := afero.NewMemMapFs()
	util.Must(archive.UntarReader(f, true, mfs, nil))
	if readFile(mfs, "zero.txt") != "zero\n" {
		t.Error("zero.txt")
	}
	if readFile(mfs, "one/one.txt") != "one\n" {
		t.Error("one.txt")
	}
}

func readFile(fs afero.Fs, path string) string {
	f, err := fs.Open(path)
	util.Must(err)
	defer f.Close()
	dat, err := ioutil.ReadAll(f)
	util.Must(err)
	return string(dat)
}

func createFile(dir, path, content string) {
	p := filepath.Join(dir, path)
	util.Must(os.MkdirAll(filepath.Dir(p), os.ModePerm))
	f, err := os.Create(p)
	util.Must(err)
	util.Must(util.PropagateCloseError(f, func() error {
		_, err := f.WriteString(content)
		return err
	}))
}
