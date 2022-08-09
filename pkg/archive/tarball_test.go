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

package archive

import (
	"io"
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
	if err := tarball.WriteFile(fs, "", "foo.txt"); err != nil {
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
	c, _ := io.ReadAll(f)
	if string(c) != "hello, world\n" {
		t.Error("bad content")
	}
}
