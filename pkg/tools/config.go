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

package tools

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	path string
	data *jnode.Node
}

func ReadConfigFile(path string) *Config {
	c := &Config{}
	d, err := ioutil.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Warnf("Could not read {warning:%s} - {warning:%s}", path, err)
		}
		return c
	}
	var m map[string]interface{}
	err = yaml.Unmarshal(d, &m)
	if err != nil {
		log.Warnf("Could not parse {warning:%s} - {warning:%s}", path, err)
	}
	c.data = jnode.FromMap(m)
	c.path = path
	return c
}
