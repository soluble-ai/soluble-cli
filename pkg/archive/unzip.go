package archive

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

var _ Unpack = Unzip

func Unzip(src afero.File, fs afero.Fs, options *Options) error {
	info, err := src.Stat()
	if err != nil {
		return err
	}
	r, err := zip.NewReader(src, info.Size())
	if err != nil {
		return err
	}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			err := fs.MkdirAll(f.Name, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}
		if err = fs.MkdirAll(filepath.Dir(f.Name), os.ModePerm); err != nil {
			return err
		}
		err = func() error {
			out, err := fs.OpenFile(f.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer out.Close()
			err = func() error {
				in, err := f.Open()
				if err != nil {
					return nil
				}
				defer in.Close()
				err = options.copy(out, in)
				if err != nil && !errors.Is(err, io.EOF) {
					return err
				}
				return nil
			}()
			return err
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
