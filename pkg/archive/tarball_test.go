package archive

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
)

func TestTarball(t *testing.T) {
	fs := afero.NewMemMapFs()
	foo, err := fs.Create("foo.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = foo.WriteString("hello, world\n")
	if err := foo.Close(); err != nil {
		t.Fatal(err)
	}
	tarball, err := NewTarballFileWriter(fs, "foo.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	if err := tarball.WriteFile(fs, "foo.txt"); err != nil {
		t.Fatal(err)
	}
	tarball.Close()
	_, _ = tarball.GetFile().Seek(0, io.SeekStart)
	if err := Untar(tarball.GetFile(), afero.NewBasePathFs(fs, "contents"), nil); err != nil {
		t.Fatal(err)
	}
	if err := tarball.GetFile().Close(); err != nil {
		t.Fatal(err)
	}
	f, err := fs.Open("contents/foo.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c, _ := ioutil.ReadAll(f)
	if string(c) != "hello, world\n" {
		t.Error("bad content")
	}
}
