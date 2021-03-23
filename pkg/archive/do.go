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
