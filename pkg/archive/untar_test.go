package archive

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestUntar(t *testing.T) {
	fs := afero.NewMemMapFs()
	tar, err := afero.NewOsFs().Open(filepath.Join("testdata", "hello.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if err := Untar(tar, fs, &Options{IgnoreSymLinks: true}); err != nil {
		t.Fatal(err)
	}
	if err := assertFileEquals(fs, "hello.txt", "hello, world\n"); err != nil {
		t.Error(err)
	}
	if err := assertFileEquals(fs, "1/2/3/three.txt", "three\n"); err != nil {
		t.Error(err)
	}
}

func TestUntarTruncate(t *testing.T) {
	fs := afero.NewMemMapFs()
	tar, err := afero.NewOsFs().Open(filepath.Join("testdata", "hello.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if err := Untar(tar, fs, &Options{TruncateFileSize: 1, IgnoreSymLinks: true}); err != nil {
		t.Fatal(err)
	}
	if err := assertFileEquals(fs, "hello.txt", "h"); err != nil {
		t.Error(err)
	}
}

func TestUntarSymlink(t *testing.T) {
	dir, err := ioutil.TempDir("", "testuntar*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	if err := Do(Untar, filepath.Join("testdata", "hello.tar.gz"), dir, nil); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(filepath.Join(dir, "three.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if (info.Mode() & os.ModeSymlink) == 0 {
		t.Error("three.txt is not a symlink")
	}
}
