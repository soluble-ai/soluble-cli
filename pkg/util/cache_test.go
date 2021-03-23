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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	assert := assert.New(t)
	fill := func(s string) interface{} { return s }
	c := NewCache(2)
	assert.Equal("hello", c.Get("hello", fill))
	assert.Equal("world", c.Get("world", fill))
	assert.Equal("hello", c.Get("hello", fill))
	assert.Equal(uint32(2), c.entries["hello"].use)
	assert.Equal(uint32(1), c.entries["world"].use)
	c.Put("blah", 5)
	assert.Equal(5, c.Get("blah", fill))
	assert.False(c.Contains("world"))
	assert.True(c.Contains("hello"))
}
