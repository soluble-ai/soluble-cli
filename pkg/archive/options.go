package archive

import (
	"errors"
	"io"
)

type Options struct {
	TruncateFileSize int64
	IgnoreSymLinks   bool
}

func (o *Options) copy(out io.Writer, in io.Reader) (err error) {
	if o != nil && o.TruncateFileSize > 0 {
		_, err = io.CopyN(out, in, o.TruncateFileSize)
		if errors.Is(err, io.EOF) {
			err = nil
		}
	} else {
		_, err = io.Copy(out, in)
	}
	return
}

func (o *Options) ignoreSymLinks() bool {
	return o != nil && o.IgnoreSymLinks
}
