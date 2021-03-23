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

package login

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
)

var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func MakeState() string {
	const n = 64
	var bb bytes.Buffer
	bb.Grow(n)
	l := uint16(len(letters))
	b := make([]byte, 2)
	for i := 0; i < n; i++ {
		_, _ = rand.Read(b)
		bb.WriteRune(letters[binary.BigEndian.Uint16(b)%l])
	}
	return bb.String()
}
