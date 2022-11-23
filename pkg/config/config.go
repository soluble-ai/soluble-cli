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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/go-homedir"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

var GlobalConfig = &struct {
	Profiles       map[string]*ProfileT
	CurrentProfile string
	ModelLocations []string `json:",omitempty"`
}{}

// Config points to the current profile
var (
	config             = &ProfileT{}
	ConfigFile         string
	ConfigDir          string
	configFileRead     string
	migrationAvailable bool
)

const Redacted = "*** redacted ***"

type ProfileT struct {
	ProfileName         string `json:"-"`
	APIServer           string
	APIToken            string `json:",omitempty"`
	TLSNoVerify         bool   `json:",omitempty"`
	Organization        string
	LaceworkProfileName string `json:",omitempty"`
	// ConfiguredAccount is the account that this profile was configured for.
	// It is set when `iac configure` runs. The need for this is non-obvious:
	// the (IAC) Organzation is tied to an account, so if the account was switched
	// (e.g. running lacework switch-profile) then the Organization is no longer
	// correct.
	ConfiguredAccount string `json:",omitempty"`
	lacework          *LaceworkProfile
}

func Get() *ProfileT {
	if GlobalConfig.Profiles == nil {
		Load()
	}
	return config
}

func SelectProfile(name string) bool {
	result := false
	config = GlobalConfig.Profiles[name]
	if config == nil {
		config = &ProfileT{
			ProfileName: name,
			APIServer:   "https://api.soluble.cloud",
		}
		GlobalConfig.Profiles[name] = config
		result = true
	}
	GlobalConfig.CurrentProfile = name
	return result
}

func CopyProfile(sourceName string) error {
	if GlobalConfig.CurrentProfile == "" {
		return fmt.Errorf("no current profile to copy to")
	}
	source := GlobalConfig.Profiles[sourceName]
	if source == nil {
		return fmt.Errorf("no source profile %s to copy from", sourceName)
	}
	copy := *source
	copy.ProfileName = GlobalConfig.CurrentProfile
	GlobalConfig.Profiles[GlobalConfig.CurrentProfile] = &copy
	SelectProfile(GlobalConfig.CurrentProfile)
	return nil
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
	p.ProfileName = rename
	delete(GlobalConfig.Profiles, name)
	if GlobalConfig.CurrentProfile == name {
		GlobalConfig.CurrentProfile = rename
	}
	log.Infof("Renamed profile {info:%s} to {info:%s}", name, rename)
	return nil
}

func (c *ProfileT) PrintableJSON() *jnode.Node {
	cfg := *config
	if cfg.APIToken != "" {
		cfg.APIToken = Redacted
	}
	var m map[string]interface{}
	dat, _ := json.Marshal(cfg)
	_ = json.Unmarshal(dat, &m)
	m["ConfigFile"] = configFileRead
	m["ProfileName"] = cfg.ProfileName
	if c.lacework != nil {
		lm := map[string]interface{}{}
		m["Lacework"] = lm
		if c.lacework.Account != "" {
			lm["account"] = c.lacework.Account
		}
		if c.lacework.APIKey != "" {
			lm["api_key"] = c.lacework.APIKey
		}
		if c.lacework.APISecret != "" {
			lm["api_secret"] = Redacted
		}
	}
	s, _ := json.Marshal(m)
	n, _ := jnode.FromJSON(s)
	return n
}

func (c *ProfileT) GetLaceworkProfile() *LaceworkProfile {
	return c.lacework
}

func (c *ProfileT) SetLaceworkProfile(name string) {
	if name == "" {
		name = "default"
	}
	c.LaceworkProfileName = name
	c.lacework = getLaceworkProfile(name)
	if c.lacework != nil {
		log.Debugf("Found lacework profile {info:%s}", name)
	}
}

func Save() error {
	dir := filepath.Dir(ConfigFile)
	if dir != "" {
		if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(dir, 0777)
			if err != nil {
				log.Errorf("Could not create config directory {info:%s}: {danger:%s}", dir, err.Error())
				return err
			}
		}
	}
	if configFileRead != "" && configFileRead != ConfigFile {
		log.Infof("Migrating legacy config file {info:%s} to {info:%s}", configFileRead, ConfigFile)
		err := os.Rename(configFileRead, ConfigFile)
		if err != nil {
			log.Errorf("Could not rename legacy config file {info:%s} to {info:%s}: {danger:%s}",
				configFileRead, ConfigFile, err.Error())
			return err
		}
	}
	log.Infof("Updating {info:%s}\n", ConfigFile)
	file, err := os.OpenFile(ConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err == nil {
		defer file.Close()
		enc := json.NewEncoder(file)
		enc.SetIndent("", "  ")
		err = enc.Encode(GlobalConfig)
	}
	if err != nil {
		log.Errorf("Failed to save {info:%s}: {danger:%s}", ConfigFile, err.Error())
	}
	return err
}

func Set(name, value string) error {
	dat, err := json.Marshal(config)
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
				err = json.Unmarshal(dat, config)
			}
		}
	}
	return err
}

// Reset loaded configuration to unloaded
func Reset() {
	configFileRead = ""
	ConfigDir = ""
	ConfigFile = ""
	laceworkProfiles = nil
	GlobalConfig.Profiles = nil
	GlobalConfig.CurrentProfile = ""
}

func Load() {
	if ConfigDir == "" {
		ConfigDir = os.Getenv("SOLUBLE_CONFIG_DIR")
		if ConfigDir == "" {
			ConfigDir, _ = homedir.Expand("~/.config/lacework")
		}
	}
	if ConfigFile == "" {
		ConfigFile = os.Getenv("SOLUBLE_CONFIG_FILE")
		if ConfigFile == "" {
			ConfigFile = filepath.Join(ConfigDir, "cli-config.json")
			if !util.FileExists(ConfigFile) {
				legacyConfigFile, _ := homedir.Expand("~/.soluble/cli-config.json")
				if util.FileExists(legacyConfigFile) {
					configFileRead = legacyConfigFile
					migrationAvailable = true
					log.Warnf("Using legacy config file {warning:%s}, use {primary:soluble config migrate} to migrate", configFileRead)
				}
			}
		}
	}
	if info, err := os.Stat(ConfigFile); err == nil && (info.Mode()&0077) != 0 {
		if runtime.GOOS != "windows" {
			if err := os.Chmod(ConfigFile, 0600); err == nil {
				log.Infof("Removed world & group permissions on config file {info:%s}", ConfigFile)
			} else {
				log.Warnf("Config file {info:%s} is world readable could not change permissions: {warning:%s}",
					ConfigFile, err.Error())
			}
		}
	}
	if configFileRead == "" {
		configFileRead = ConfigFile
	}
	GlobalConfig.Profiles = map[string]*ProfileT{}
	dat, err := os.ReadFile(configFileRead)
	if err != nil {
		configFileRead = ""
	} else {
		_ = json.Unmarshal(dat, GlobalConfig)
		for name, profile := range GlobalConfig.Profiles {
			profile.ProfileName = name
			profile.lacework = getLaceworkProfile(profile.LaceworkProfileName)
		}
	}
	config = GlobalConfig.Profiles[GlobalConfig.CurrentProfile]
	if config == nil {
		SelectProfile("default")
	}
	if config.ProfileName == "" {
		config.ProfileName = GlobalConfig.CurrentProfile
	}
}

func GetModelLocations() []string {
	return GlobalConfig.ModelLocations
}

func Migrate() error {
	if !migrationAvailable {
		log.Infof("Config file {info:%s} is already up-to-date", ConfigFile)
		return nil
	}
	return Save()
}
