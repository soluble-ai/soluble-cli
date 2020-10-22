package archive

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestUnzip(t *testing.T) {
	fs, err := unzipHello(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertFileEquals(fs, "README.txt", "hello, world\n"); err != nil {
		t.Error(err)
	}
	if err := assertFileEquals(fs, "foo/bar/1.txt", "one\n"); err != nil {
		t.Error(err)
	}
}

func TestUnzipTruncate(t *testing.T) {
	fs, err := unzipHello(&Options{TruncateFileSize: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := assertFileEquals(fs, "foo/bar/1.txt", "o"); err != nil {
		t.Error(err)
	}
}

func TestUnzipTree(t *testing.T) {
	dir, err := ioutil.TempDir("", "unziptree*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	three := filepath.Join(dir, "one", "two", "three")
	if err := os.MkdirAll(three, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	_ = Do(Unzip, filepath.Join("testdata", "tree.zip"), three, nil)
	_, err = os.Stat(filepath.Join(dir, "zero.txt"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Error("wrote zero.txt")
	}
}

func unzipHello(options *Options) (afero.Fs, error) {
	fs := afero.NewMemMapFs()
	err := func() error {
		in, err := afero.NewOsFs().Open(filepath.Join("testdata", "hello.zip"))
		if err != nil {
			return err
		}
		defer in.Close()
		return Unzip(in, fs, options)
	}()
	return fs, err
}

func assertFileEquals(fs afero.Fs, path, content string) error {
	in, err := fs.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	dat, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}
	s := string(dat)
	if s != content {
		return fmt.Errorf("for %s %#v != %#v", path, s, content)
	}
	return nil
}
