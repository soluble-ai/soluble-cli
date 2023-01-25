package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type AtomicFileWriter struct {
	Temp *os.File
	path string
}

var _ io.WriteCloser = (*AtomicFileWriter)(nil)

// Creates a file writer that writes to a temporary file and can
// rename it to a target file when done.  The temp file is created
// in the same directory as the target file; the rename operation
// should be atmoic in this case.
func NewAtomicFileWriter(path string) (*AtomicFileWriter, error) {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, fmt.Sprintf("%s*", filepath.Base(path)))
	if err != nil {
		return nil, err
	}
	return &AtomicFileWriter{
		Temp: tmp,
		path: path,
	}, nil
}

func (a *AtomicFileWriter) Write(p []byte) (int, error) {
	return a.Temp.Write(p)
}

// Closes and renames the temp file.
func (a *AtomicFileWriter) Rename() error {
	if err := a.Temp.Close(); err != nil {
		return err
	}
	tmp := a.Temp
	a.Temp = nil
	if err := os.Rename(tmp.Name(), a.path); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return nil
}

// Closes and deletes the tempfile unless it has already been renamed.
func (a *AtomicFileWriter) Close() error {
	if a.Temp == nil {
		return nil
	}
	if err := a.Temp.Close(); err != nil {
		return err
	}
	if err := os.Remove(a.Temp.Name()); err != nil {
		return err
	}
	a.Temp = nil
	return nil
}
