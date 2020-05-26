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

func TestStringSet(t *testing.T) {
	s := NewStringSet().AddAll("1", "two", "three")
	if s.Add("1") != false {
		t.Error(s)
	}
	if s.Add("one") != true {
		t.Error(s)
	}
	v := s.Values()
	if len(v) != 4 || v[0] != "1" || v[1] != "two" || v[3] != "one" {
		t.Error(v)
	}
	if s.Contains("one") != true || s.Contains("four") == true {
		t.Error(s)
	}
}
