package archive

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"path/filepath"

	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/afero"
)

type TarballWriter struct {
	file afero.File
	gzip *gzip.Writer
	tar  *tar.Writer
}

func NewTarballFileWriter(fs afero.Fs, path string) (*TarballWriter, error) {
	f, err := fs.Create(path)
	if err != nil {
		return nil, err
	}
	return NewTarballWriter(f), nil
}

func NewTarballWriter(file afero.File) *TarballWriter {
	t := &TarballWriter{
		file: file,
	}
	t.gzip = gzip.NewWriter(t.file)
	t.tar = tar.NewWriter(t.gzip)
	return t
}

func (t *TarballWriter) GetFile() afero.File {
	return t.file
}

func (t *TarballWriter) WriteFile(fs afero.Fs, dir, path string) error {
	var err error
	name := path
	if dir != "" {
		if filepath.IsAbs(path) {
			name, err = filepath.Rel(dir, path)
			if err != nil {
				return err
			}
		} else {
			path = filepath.Join(dir, path)
		}
	}
	f, err := fs.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	return t.Write(name, info.Size(), f)
}

func (t *TarballWriter) Write(name string, size int64, r io.Reader) error {
	h := &tar.Header{
		Name: name,
		Mode: 0666,
		Size: size,
	}
	if err := t.tar.WriteHeader(h); err != nil {
		return err
	}
	if _, err := io.Copy(t.tar, r); err != nil {
		return err
	}
	return nil
}

func (t *TarballWriter) Close() error {
	return util.CloseAll(t.tar, t.gzip)
}
