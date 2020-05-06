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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"gopkg.in/yaml.v2"
)

var GlobalConfig = &struct {
	Profiles       map[string]*ProfileT
	CurrentProfile string
	ModelLocations []string
}{}

// Config points to the current profile
var Config = &ProfileT{}
var ConfigFile string

const Redacted = "*** redacted ***"

type ProfileT struct {
	APIServer        string
	APIToken         string
	TLSNoVerify      bool
	Organization     string
	Email            string
	DefaultClusterID string
}

func SelectProfile(name string) bool {
	result := false
	Config = GlobalConfig.Profiles[name]
	if Config == nil {
		Config = &ProfileT{
			APIServer: "https://api.soluble.cloud",
		}
		GlobalConfig.Profiles[name] = Config
		result = true
	}
	GlobalConfig.CurrentProfile = name
	return result
}

func DeleteProfile(name string) bool {
	if GlobalConfig.Profiles[name] == nil {
		log.Warnf("Deleting non-existent profile {warning:%s} has no effect", name)
		return false
	}
	delete(GlobalConfig.Profiles, name)
	log.Infof("Deleted profile {info:%s}", name)
	if len(GlobalConfig.Profiles) == 0 {
		SelectProfile("default")
	} else if name == GlobalConfig.CurrentProfile {
		for p := range GlobalConfig.Profiles {
			SelectProfile(p)
			break
		}
	}
	return true
}

func RenameProfile(name, rename string) error {
	p := GlobalConfig.Profiles[name]
	if p == nil {
		return fmt.Errorf("cannot rename non-existent profile %s", name)
	}
	GlobalConfig.Profiles[rename] = p
	delete(GlobalConfig.Profiles, name)
	if GlobalConfig.CurrentProfile == name {
		GlobalConfig.CurrentProfile = rename
	}
	log.Infof("Renamed profile {info:%s} to {info:%s}", name, rename)
	return nil
}

func (c *ProfileT) String() string {
	cfg := *Config
	if cfg.APIToken != "" {
		cfg.APIToken = Redacted
	}
	s, _ := yaml.Marshal(cfg)
	return string(s)
}

func Save() {
	log.Infof("Updating {info:%s}\n", ConfigFile)
	file, err := os.Create(ConfigFile)
	if err == nil {
		defer file.Close()
		enc := json.NewEncoder(file)
		enc.SetIndent("", "  ")
		err = enc.Encode(GlobalConfig)
	}
	if err != nil {
		log.Errorf("Failed to save {info:%s}: {danger:%s}", ConfigFile, err.Error())
	}
}

func Set(name, value string) error {
	dat, err := json.Marshal(Config)
	if err == nil {
		m := map[string]interface{}{}
		err = json.Unmarshal(dat, &m)
		if err == nil {
			if name == "tlsnoverify" {
				// hack
				m[name] = value == "true"
			} else {
				m[name] = value
			}
			dat, err = json.Marshal(m)
			if err == nil {
				err = json.Unmarshal(dat, Config)
			}
		}
	}
	return err
}

func Load() {
	if ConfigFile == "" {
		ConfigFile = os.Getenv("SOLUBLE_CONFIG_FILE")
		if ConfigFile == "" {
			ConfigFile, _ = homedir.Expand("~/.soluble_cli")
		}
	}
	dat, err := ioutil.ReadFile(ConfigFile)
	if err == nil {
		_ = json.Unmarshal(dat, GlobalConfig)
	}
	if GlobalConfig.Profiles == nil {
		GlobalConfig.Profiles = map[string]*ProfileT{}
	}
	Config = GlobalConfig.Profiles[GlobalConfig.CurrentProfile]
	if Config == nil {
		SelectProfile("default")
	}
}

func UpdateFromServerProfile(result *jnode.Node) bool {
	changed := setIfChanged(&Config.Organization, result.Path("currentOrgId").AsText())
	changed = setIfChanged(&Config.Email, result.Path("email").AsText()) || changed
	return changed
}

func setIfChanged(c *string, v string) bool {
	if *c != v {
		*c = v
		return true
	}
	return false
}

func GetModelLocations() []string {
	return GlobalConfig.ModelLocations
}
