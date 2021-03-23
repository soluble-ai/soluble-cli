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
