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

import "testing"

var containsTestCases = []struct {
	slice []string
	s     string
	r     bool
}{
	{[]string{"one", "two", "three"}, "one", true},
	{[]string{"one", "two", "three"}, "two", true},
	{[]string{"one", "two", "three"}, "fower", false},
	{nil, "fiver", false},
}

func TestContains(t *testing.T) {
	for _, tc := range containsTestCases {
		if StringSliceContains(tc.slice, tc.s) != tc.r {
			t.Error(tc.slice, tc.s, tc.r)
		}
	}
}

func TestNormalizedPath(t *testing.T) {
	d := "."
	np, err := NormalizePath(d)
	if err != nil {
		t.Error(err)
	}

	if np == "." {
		t.Error("path should be absolute")
	}

	d = ""
	np, _ = NormalizePath(d)
	if len(np) == 0 {
		t.Error("should be resolved to current directory")
	}
}
