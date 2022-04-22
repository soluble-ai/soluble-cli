package capture

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

const defaultMemoryLimit = 512 * 1024

type Capture struct {
	lock        sync.Mutex
	stopped     bool
	buf         *bytes.Buffer
	file        *os.File
	MemoryLimit int
}

var _ io.WriteCloser = (*Capture)(nil)

func NewCapture() *Capture {
	return &Capture{
		buf:         &bytes.Buffer{},
		MemoryLimit: defaultMemoryLimit,
	}
}

func (cap *Capture) Write(p []byte) (int, error) {
	cap.lock.Lock()
	defer cap.lock.Unlock()
	if cap.stopped {
		return len(p), nil
	}
	if cap.buf != nil {
		if cap.buf.Len()+len(p) < cap.MemoryLimit {
			return cap.buf.Write(p)
		}
		f, err := os.CreateTemp("", "clicapture*")
		if err != nil {
			return 0, fmt.Errorf("could not create overflow capture file - %w", err)
		}
		// unlink the temp file immediately
		_ = os.Remove(f.Name())
		dat := cap.buf.Bytes()
		cap.buf = nil
		if _, err := f.Write(dat); err != nil {
			return 0, fmt.Errorf("could not write initial contents of overflow capture file - %w", err)
		}
		cap.file = f
	}
	if cap.file != nil {
		return cap.file.Write(p)
	}
	return 0, fmt.Errorf("cannot capture further output")
}

func (cap *Capture) Close() (err error) {
	cap.lock.Lock()
	defer cap.lock.Unlock()
	if cap.file != nil {
		err = cap.file.Close()
	}
	cap.file = nil
	cap.buf = nil
	return err
}

func (cap *Capture) Output() (io.ReadCloser, error) {
	cap.lock.Lock()
	defer cap.lock.Unlock()
	cap.stopped = true
	if cap.buf != nil {
		dat := cap.buf.Bytes()
		cap.buf = nil
		return &noopReadCloser{bytes.NewReader(dat)}, nil
	}
	if cap.file != nil {
		f := cap.file
		cap.file = nil
		_, err := f.Seek(0, 0)
		if err != nil {
			_ = f.Close()
			return nil, err
		}
		return f, nil
	}
	return nil, fmt.Errorf("capture output has failed")
}

func (cap *Capture) OutputBytes() ([]byte, error) {
	r, err := cap.Output()
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}

type noopReadCloser struct{ io.Reader }

func (*noopReadCloser) Close() error { return nil }
