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
	"regexp"
	"strconv"
)

var Version string = "<unknown>"
var BuildTime string = ""

var versionRe = regexp.MustCompile(`([0-9]+)(\.([0-9]+)?(\.([0-9]+)(.*)?)?)?`)

const (
	Major = 1
	Minor = 3
	Patch = 5
	Qual  = 6
)

func IsCompatible(version string) bool {
	return isCompatible(Version, version)
}

func isCompatible(v1, v2 string) bool {
	m1 := versionRe.FindStringSubmatch(v1)
	m2 := versionRe.FindStringSubmatch(v2)
	if m1 == nil || m2 == nil {
		return false
	}
	if m1[Major] != m2[Major] {
		return false
	}
	if m2[Minor] == "" {
		return true
	}
	mi1 := n(m1[Minor])
	mi2 := n(m2[Minor])
	if mi1 != mi2 {
		return mi1 > mi2
	}
	q1 := m1[Qual]
	p1 := n(m1[Patch])
	if q1 != "" {
		// if v1 has a qualifier consider it compatible with the next patch
		// e.g. 0.3.7-foo is compatible with 0.3.8
		p1 += 1
	}
	p2 := n(m2[Patch])
	return p1 >= p2
}

func n(s string) int {
	i, err := strconv.Atoi(s)
	if err == nil {
		return i
	}
	return -1
}
