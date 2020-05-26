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

type StringSet struct {
	values []string
	set    map[string]interface{}
}

func NewStringSet() *StringSet {
	return &StringSet{
		set: map[string]interface{}{},
	}
}

func (ss *StringSet) Contains(s string) bool {
	_, ok := ss.set[s]
	return ok
}

// Adds s to the set and returns true if the string wasn't
// already present
func (ss *StringSet) Add(s string) bool {
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
