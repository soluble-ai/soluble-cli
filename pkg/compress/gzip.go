package compress

import (
	"compress/gzip"
	"io"
)

type gzipPipe struct {
	pr  *io.PipeReader
	pw  *io.PipeWriter
	gzw *gzip.Writer
	err error
}

func (gz *gzipPipe) copy(src io.Reader) {
	_, gz.err = io.Copy(gz.gzw, src)
	gz.err = or(gz.err, gz.gzw.Close())
	gz.err = or(gz.err, gz.pw.Close())
}

func (gz *gzipPipe) Close() error {
	if err := gz.pr.Close(); err != nil {
		return err
	}
	return gz.err
}

func (gz *gzipPipe) Read(p []byte) (int, error) {
	return gz.pr.Read(p)
}

func NewGZIPPipe(src io.Reader) io.ReadCloser {
	r, w := io.Pipe()
	gz := &gzipPipe{
		pr:  r,
		pw:  w,
		gzw: gzip.NewWriter(w),
	}
	go gz.copy(src)
	return gz
}

func or(e1 error, e2 error) error {
	if e1 != nil {
		return e1
	}
	return e2
}
