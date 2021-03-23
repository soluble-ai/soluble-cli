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
