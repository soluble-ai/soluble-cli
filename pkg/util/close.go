// Copyright 2021 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import "io"

// CloseAll calls Close() on every closer, returning the err from
// the first one that returns a non-nil error.
func CloseAll(closers ...io.Closer) error {
	var result error
	for _, c := range closers {
		if c != nil {
			err := c.Close()
			if result == nil {
				result = err
			}
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
