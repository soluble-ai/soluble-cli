package archive

import (
	"archive/tar"
	"compress/gzip"
	"io"

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

func (t *TarballWriter) WriteFile(fs afero.Fs, path string) error {
	f, err := fs.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Write(f)
}

func (t *TarballWriter) Write(file afero.File) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	h := &tar.Header{
		Name: file.Name(),
		Mode: int64(info.Mode()),
		Size: info.Size(),
	}
	if err := t.tar.WriteHeader(h); err != nil {
		return err
	}
	if _, err := io.Copy(t.tar, file); err != nil {
		return err
	}
	return nil
}

func (t *TarballWriter) Close() error {
	return util.CloseAll(t.tar, t.gzip)
}
