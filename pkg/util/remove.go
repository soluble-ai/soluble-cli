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

package util

import "github.com/soluble-ai/go-jnode"

func RemoveJNodeElementsIf(n *jnode.Node, f func(*jnode.Node) bool) *jnode.Node {
	// don't allocate a new array if the filter doesn't exclude anything
	var result *jnode.Node
	for i, e := range n.Elements() {
		if f(e) {
			if result == nil {
				result = jnode.NewArrayNode()
				for j := 0; j < i; j++ {
					result.Append(n.Get(j))
				}
			}
		} else if result != nil {
			result.Append(e)
		}
	}
	if result != nil {
		return result
	}
	return n
}

func RemoveJNodeEntriesIf(n *jnode.Node, f func(key string, value *jnode.Node) bool) *jnode.Node {
	for k, v := range n.Entries() {
		if f(k, v) {
			n.Remove(k)
		}
	}
	return n
}
