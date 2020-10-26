package util

import "io"

// CloseAll calls Close() on every closer, returning the err from
// the first one that returns a non-nil error.
func CloseAll(closers ...io.Closer) error {
	var result error
	for _, c := range closers {
		err := c.Close()
		if result == nil {
			result = err
		}
	}
	return result
}

// PropagateCloseError calls f then calls closer.Close(), returning the error from
// Close() if f did not return an error.  Errors from Close() when writing can return
// important errors that should not be thrown away.
func PropagateCloseError(closer io.Closer, f func() error) (err error) {
	err = f()
	cerr := closer.Close()
	if err == nil {
		err = cerr
	}
	return err
}
