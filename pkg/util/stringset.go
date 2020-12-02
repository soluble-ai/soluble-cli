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

import "encoding/json"

type StringSet struct {
	values []string
	set    map[string]interface{}
}

func NewStringSet() *StringSet {
	return &StringSet{
		set: map[string]interface{}{},
	}
}

func NewStringSetWithValues(values []string) *StringSet {
	s := NewStringSet()
	for _, value := range values {
		s.Add(value)
	}
	return s
}

func (ss *StringSet) Contains(s string) bool {
	_, ok := ss.set[s]
	return ok
}

// Adds s to the set and returns true if the string wasn't
// already present
func (ss *StringSet) Add(s string) bool {
	if ss.set == nil {
		ss.set = map[string]interface{}{}
	}
	_, ok := ss.set[s]
	if !ok {
		ss.set[s] = nil
		ss.values = append(ss.values, s)
	}
	return !ok
}

func (ss *StringSet) AddAll(values ...string) *StringSet {
	for _, s := range values {
		ss.set[s] = nil
	}
	ss.values = append(ss.values, values...)
	return ss
}

func (ss *StringSet) Values() []string {
	return ss.values
}

func (ss *StringSet) Len() int {
	return len(ss.set)
}

func (ss *StringSet) Reset() {
	ss.set = nil
	ss.values = nil
}

func (ss *StringSet) Get(i int) string {
	return ss.values[i]
}

func (ss *StringSet) MarshalJSON() ([]byte, error) {
	if ss.values != nil {
		return json.Marshal(ss.values)
	}
	return []byte(`[]`), nil
}
