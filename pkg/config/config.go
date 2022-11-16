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
	"strings"

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
	Config             = &ProfileT{}
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

func SelectProfile(name string) bool {
	result := false
	Config = GlobalConfig.Profiles[name]
	if Config == nil {
		Config = &ProfileT{
			ProfileName: name,
			APIServer:   "https://api.soluble.cloud",
		}
		GlobalConfig.Profiles[name] = Config
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
	cfg := *Config
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

func (c *ProfileT) GetAppURL() string {
	const httpAPI = "https://api."
	apiServer := c.GetAPIServer()
	if strings.HasPrefix(apiServer, httpAPI) {
		return "https://app." + apiServer[len(httpAPI):]
	}
	return "https://app.soluble.cloud"
}

func (c *ProfileT) GetAPIToken() string {
	token := strings.TrimSpace(os.Getenv("SOLUBLE_API_TOKEN"))
	if token != "" {
		return token
	}
	return c.APIToken
}

func (c *ProfileT) GetOrganization() string {
	org := os.Getenv("LW_IAC_ORGANIZATION")
	if org != "" {
		return org
	}
	return c.Organization
}

func (c *ProfileT) AssertAPITokenFromConfig() error {
	if os.Getenv("SOLUBLE_API_TOKEN") != "" {
		return fmt.Errorf("the environment variable SOLUBLE_API_TOKEN is set")
	}
	return nil
}

func (c *ProfileT) GetAPIServer() string {
	server := strings.TrimSpace(os.Getenv(("SOLUBLE_API_SERVER")))
	if server != "" {
		return server
	}
	server = os.Getenv("LW_IAC_API_URL")
	if server != "" {
		return server
	}
	return c.APIServer
}

func (c *ProfileT) GetLaceworkProfile() *LaceworkProfile {
	return c.lacework
}

func (c *ProfileT) GetLaceworkAccount() string {
	account := os.Getenv("LW_ACCOUNT")
	if account != "" {
		return account
	}
	if c.lacework != nil {
		return c.lacework.Account
	}
	return ""
}

func (c *ProfileT) GetLaceworkAPIKey() string {
	key := os.Getenv("LW_API_KEY")
	if key != "" {
		return key
	}
	if c.lacework != nil {
		return c.lacework.APIKey
	}
	return ""
}

func (c *ProfileT) GetLaceworkAPISecret() string {
	secret := os.Getenv("LW_API_SECRET")
	if secret != "" {
		return secret
	}
	if c.lacework != nil {
		return c.lacework.APISecret
	}
	return ""
}

// Verify that the configuration chosen is usuable
// In particular, verify that LW_ACCOUNT is the same, because otherwise Organization
// is likely to be wrong.
func (c *ProfileT) ValidateConfiguration() error {
	// Try and detect a couple of situations where the IAC component
	// is not configured properly and tell the user what to do
	account := c.GetLaceworkAccount()
	legacyToken := c.GetAPIToken()
	if !IsRunningAsComponent() {
		if legacyToken != "" && account == "" {
			log.Infof("Using legacy soluble authentication")
			return nil
		}
	}
	if legacyToken == "" && account == "" {
		if c.ConfiguredAccount != "" {
			log.Errorf("The IAC profile {info:%s} cannot be used when invoking the IAC component directly.", c.ProfileName)
			log.Infof("Run the command using the lacework CLI instead with {primary:lacework iac ...}")
			return fmt.Errorf("must use 'lacework iac ...'")
		}
		// We have no configuration at all.
		log.Infof("The IAC profile {info:%s} must be configured by running {primary:%s configure}", c.ProfileName, CommandInvocation())
		return fmt.Errorf("configuration required")
	}
	if account != "" {
		if sub := os.Getenv("LW_SUBACCOUNT"); sub != "" {
			// Neither the cli nor api-server support subaccounts currently
			log.Errorf("The IAC component does not support subaccounts")
			return fmt.Errorf("configuration unsupported")
		}
		// We've got a lacework account, so we either need to be linked to a
		// lacework profile (and that profile must exist), or the account
		// must match the account we've been configured with.
		if c.LaceworkProfileName != "" {
			if c.lacework == nil {
				log.Errorf("The IAC profile {primary:%s} is configured to use the lacework profile {primary:%s}",
					c.ProfileName, c.LaceworkProfileName)
				log.Errorf("but that profile does not exist.")
				return fmt.Errorf("lacework profile %s does not exist", c.LaceworkProfileName)
			}
			if account != c.lacework.Account {
				log.Errorf("The IAC profile {info:%s} is configured to use account {info:%s} but is running with account {info:%s}",
					c.ProfileName, c.lacework.Account, account)
				log.Errorf("Run {primary:%s configure --reconfigure} to change configuration, or use {primary:%s configure --profile <new-name>} to create a new profile for this account",
					CommandInvocation(), CommandInvocation())
				return fmt.Errorf("configuration is invalid")
			}
			log.Infof("Using lacework authentication for account {info:%s} linked to lacework profile {info:%s}",
				account, c.LaceworkProfileName)
			return nil
		}
		if c.ConfiguredAccount == "" {
			log.Errorf("The IAC profile {info:%s} has not been configured, please run {primary:%s configure}", c.ProfileName, CommandInvocation())
			return fmt.Errorf("configuration required")
		}
		if account != c.ConfiguredAccount {
			log.Errorf("The IAC profile {primary:%s} is configured to use account {info:%s} but is running with {info:%s}",
				c.ProfileName, c.ConfiguredAccount, account)
			log.Errorf("Run {primary:%s configure --reconfigure} to change configuration, or use {primary:%s configure --profile <new-name>} to create a new profile for this account",
				CommandInvocation(), CommandInvocation())
			return fmt.Errorf("configuration is invalid")
		}
		log.Infof("Using lacework authentication for account {primary:%s} IAC organization {info:%s}",
			c.ConfiguredAccount, c.GetOrganization())
		return nil
	}
	return nil
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
	Config = GlobalConfig.Profiles[GlobalConfig.CurrentProfile]
	if Config == nil {
		SelectProfile("default")
	}
	if Config.ProfileName == "" {
		Config.ProfileName = GlobalConfig.CurrentProfile
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
