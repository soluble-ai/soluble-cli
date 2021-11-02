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

package fingerprint

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

// This is a straightforward translation of github's partial fingerprint
// algorithm into Go.  The original version (MIT license) is here:
// https://github.com/github/codeql-action/blob/3d63fa4dad131d9f66844df65a14115b5af6afeb/src/fingerprints.ts#L129

const mod = uint64(37)
const blockSize = 100

var firstMod uint64

func updateHash(window []rune, current rune, index int, hashRaw uint64) (int, uint64) {
	begin := window[index]
	window[index] = current
	hashRaw = mod*hashRaw + uint64(current) - firstMod*uint64(begin)
	index = (index + 1) % blockSize
	return index, hashRaw
}

func outputHash(hashCounts map[uint64]int, hashRaw uint64, lineNumbers []int, index int, lineFunc func(int, string)) {
	hashCounts[hashRaw]++
	hashValue := fmt.Sprintf("%016x:%d", hashRaw, hashCounts[hashRaw])
	lineFunc(lineNumbers[index], hashValue)
	lineNumbers[index] = -1
}

// Compute partial fingerprints for each line in a file
func Partial(r *bufio.Reader, lineFunc func(int, string)) error {
	var (
		window      = make([]rune, blockSize)
		lineNumbers = make([]int, blockSize)
		hashRaw     uint64
		index       int
		lineNumber  int
		lineStart   = true
		prevCR      bool
		hashCounts  = map[uint64]int{}
	)
	for i := range lineNumbers {
		lineNumbers[i] = -1
	}
	for {
		current, _, err := r.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		if current == ' ' || current == '\t' || (prevCR && current == '\n') {
			continue
		}
		if current == '\r' {
			current = '\n'
			prevCR = true
		} else {
			prevCR = false
		}
		if lineNumbers[index] != -1 {
			outputHash(hashCounts, hashRaw, lineNumbers, index, lineFunc)
		}
		if lineStart {
			lineStart = false
			lineNumber++
			lineNumbers[index] = lineNumber
		}
		if current == '\n' {
			lineStart = true
		}
		index, hashRaw = updateHash(window, current, index, hashRaw)
	}
	for i := 0; i < blockSize; i++ {
		if lineNumbers[index] != -1 {
			outputHash(hashCounts, hashRaw, lineNumbers, index, lineFunc)
		}
		index, hashRaw = updateHash(window, 0, index, hashRaw)
	}
	return nil
}

func init() {
	firstMod = 1
	for i := 0; i < blockSize; i++ {
		firstMod *= mod
	}
}
