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

package version

import (
	"fmt"
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("missing version")
	}
	if Version != "<unknown>" && !IsCompatible(Version) {
		t.Error(Version)
	}
}

var compatTestCases = []struct {
	v1, v2 string
	compat bool
}{
	{"0.4.0", "0.3.11", true},
	{"1.1.0", "1", true},
	{"2.10.7", "1", false},
	{"2.10.7", "2.9", true},
	{"2.10.7", "2.10.7", true},
	{"2.10.7", "2.10.6", true},
	{"2.10.7", "2.10.8", false},
	{"2.9.7", "2.10.8", false},
	{"0.3.9-foo", "0.3.10", true},
}

func TestCompat(t *testing.T) {
	for _, tc := range compatTestCases {
		if v := isCompatible(tc.v1, tc.v2); v != tc.compat {
			t.Error(tc.v1, tc.v2, fmt.Sprintf("%v != %v", v, tc.compat))
		}
	}
}
