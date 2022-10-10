package config

import (
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml/v2"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type LaceworkProfile struct {
	Account   string `toml:"account"`
	APIKey    string `toml:"api_key"`
	APISecret string `toml:"api_secret"`
}

var laceworkConfigFile string
var laceworkProfiles map[string]*LaceworkProfile

func getLaceworkProfile(laceworkProfileName string, cliProfileName string) *LaceworkProfile {
	if laceworkProfiles == nil {
		if laceworkConfigFile == "" {
			laceworkConfigFile, _ = homedir.Expand("~/.lacework.toml")
		}
		f, err := os.Open(laceworkConfigFile)
		if err == nil {
			defer f.Close()
			dec := toml.NewDecoder(f)
			_ = dec.Decode(&laceworkProfiles)
			log.Debugf("Loaded lacework config file {primary:%s}", laceworkConfigFile)
		}
	}
	name := laceworkProfileName
	if name == "" {
		name = cliProfileName
	}
	if name == "" {
		name = "default"
	}
	return laceworkProfiles[name]
}
