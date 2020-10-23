package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

var _ Unpack = Untar

func UntarReader(r io.Reader, compressed bool, fs afero.Fs, options *Options) error {
	if compressed {
		gunzip, err := gzip.NewReader(r)
		if err != nil {
			return err
		}
		defer gunzip.Close()
		r = gunzip
	}
	t := tar.NewReader(r)
	for {
		header, err := t.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := fs.MkdirAll(header.Name, os.ModePerm); err != nil {
				return err
			}
		case tar.TypeReg:
			err := func() error {
				dir := filepath.Dir(header.Name)
				if dir != "." {
					if err := fs.MkdirAll(dir, os.ModePerm); err != nil {
						return err
					}
				}
				f, err := fs.OpenFile(header.Name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
				if err != nil {
					return err
				}
				defer f.Close()
				if err := options.copy(f, t); err != nil {
					return err
				}
				return nil
			}()
			if err != nil {
				return err
			}
		case tar.TypeSymlink:
			if options.ignoreSymLinks() {
				break
			}
			sfs, ok := fs.(afero.Symlinker)
			if !ok {
				return fmt.Errorf("this filesystem does not support symlinks")
			}
			if err := sfs.SymlinkIfPossible(header.Linkname, header.Name); err != nil {
				return err
			}
		default:
			// just ignore anything else
		}
	}
	return nil
}

func Untar(src afero.File, fs afero.Fs, options *Options) error {
	return UntarReader(src, strings.HasSuffix(src.Name(), ".gz"), fs, options)
}
