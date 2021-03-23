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

import "testing"

func TestMakeState(t *testing.T) {
	s1 := MakeState()
	if len(s1) != 64 {
		t.Error(s1)
	}
	if s2 := MakeState(); len(s2) != 64 || s1 == s2 {
		t.Error(s1, s2)
	}
}
