package archive

import (
	"errors"
	"os"

	"github.com/spf13/afero"
)

type Unpack func(src afero.File, fs afero.Fs, options *Options) error

// Unpack an archive in the real OS filesystem.  The destination
// directory will be created fresh.
func Do(unpack Unpack, archive, dest string, options *Options) error {
	fs := afero.NewOsFs()
	f, err := fs.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := fs.RemoveAll(dest); err != nil {
		return err
	}
	if err := fs.MkdirAll(dest, os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	return unpack(f, afero.NewBasePathFs(fs, dest), options)
}
