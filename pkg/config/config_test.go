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

package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	assert := assert.New(t)
	f, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	ConfigFile = f.Name()
	f.Close()
	os.Remove(ConfigFile)
	Load()
	config.APIToken = "xxx"
	assert.NoError(Save())
	defer os.Remove(ConfigFile)
	Load()
	if config.APIToken != "xxx" {
		t.Error(GlobalConfig)
	}
	SelectProfile("test")
	if GlobalConfig.CurrentProfile != "test" || config.APIToken != "" {
		t.Error(GlobalConfig)
	}
	config.APIToken = "yyy"
	assert.NoError(Save())
	Load()
	if GlobalConfig.CurrentProfile != "test" || config.APIToken != "yyy" {
		t.Error(GlobalConfig)
	}
	n := config.PrintableJSON()
	if strings.Contains(n.Path("APIToken").AsText(), "yyy") {
		t.Error(config)
	}
	_ = Set("tlsnoverify", "true")
	if !config.TLSNoVerify {
		t.Error(config)
	}
	DeleteProfile("test")
	if _, ok := GlobalConfig.Profiles["test"]; ok {
		t.Error("profile wasn't deleted")
	}
}
