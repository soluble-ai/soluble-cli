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

package model

type ContextValueSupplier func(string) (string, error)

var contextValueSuppliers = map[string]ContextValueSupplier{}

type ContextValues struct {
	values map[string]string
}

func AddContextValueSupplier(name string, supplier ContextValueSupplier) {
	contextValueSuppliers[name] = supplier
}

func (c *ContextValues) Get(name string) (string, error) {
	supplier := contextValueSuppliers[name]
	if supplier != nil {
		return supplier(name)
	}
	return c.values[name], nil
}

func (c *ContextValues) Set(name, value string) {
	if c.values == nil {
		c.values = map[string]string{}
	}
	c.values[name] = value
}
