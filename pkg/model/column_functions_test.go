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

package model

import (
	"testing"

	"github.com/soluble-ai/go-jnode"
	"github.com/stretchr/testify/assert"
)

func TestParseColumnFunction(t *testing.T) {
	var testCases = []struct {
		s    string
		name string
		args []string
	}{
		{"max(foo)", "max", []string{"foo"}},
		{"min", "min", []string{}},
		{"cons(first,second)", "cons", []string{"first", "second"}},
	}
	for _, c := range testCases {
		name, args := ColumnFunctionType(c.s).parse()
		if name != c.name {
			t.Error(name, c.name)
		}
		if len(args) != len(c.args) {
			t.Error(len(args), len(c.args))
		} else {
			for i := range args {
				if args[i] != c.args[i] {
					t.Error(args, c.args)
				}
			}
		}
	}
}

func TestMinMaxRange(t *testing.T) {
	var testCases = []struct {
		name   string
		a, b   interface{}
		result interface{}
	}{
		{"min", "5", "10", 5.},
		{"max", -5, "100", 100.},
		{"range", -1, 10, "-1 - 10"},
	}
	assert := assert.New(t)
	for _, tc := range testCases {
		cf, err := minMaxColumnFunctions(tc.name, []string{"v"})
		assert.Nil(err)
		assert.NotNil(cf)
		n := jnode.NewObjectNode()
		n.PutArray("v").Append(tc.a).Append(tc.b)
		v := cf(n)
		assert.Equal(tc.result, v)
	}
}
