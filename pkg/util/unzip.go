// Copyright 2020 Soluble Inc
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

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	securejoin "github.com/cyphar/filepath-securejoin"
)

// Unzip will decompress a zip archived file,
// copying all files and folders
// within the zip file (parameter 1)
// to an output directory (parameter 2).
func Unzip(src string, destination string) ([]string, error) {
	var filenames []string
	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}

	defer r.Close()
	for _, f := range r.File {
		fpath, err := securejoin.SecureJoin(destination, f.Name)
		if err != nil {
			return nil, err
		}
		// Checking for any invalid file paths
		if !strings.HasPrefix(fpath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s is an illegal filepath", fpath)
		}
		filenames = append(filenames, fpath)
		if f.FileInfo().IsDir() {
			// Creating a new Folder
			err := os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return filenames, err
			}
			continue
		}

		// Creating the files in the target directory
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		// The created file will be stored in
		// outFile with permissions to write &/or truncate
		outFile, err := os.OpenFile(fpath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			f.Mode())
		if err != nil {
			// with filenames gathered so far
			// and err message
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			// with filenames gathered so far
			// and err message back to main function
			return filenames, err
		}

		_, err = io.CopyN(outFile, rc, 50<<(10*2)) // 50MB Max copy size, must handle EOF
		outFile.Close()
		rc.Close()
		if err != nil && err != io.EOF {
			return filenames, err
		}
	}
	return filenames, nil
}
