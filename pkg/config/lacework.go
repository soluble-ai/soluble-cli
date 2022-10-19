package config

import (
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml/v2"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type LaceworkProfile struct {
	Name      string `toml:"-"`
	Account   string `toml:"account"`
	APIKey    string `toml:"api_key"`
	APISecret string `toml:"api_secret"`
}

var (
	laceworkProfiles map[string]*LaceworkProfile
)

func LoadLaceworkProfiles(configFile string) {
	// exposed for testing
	if configFile == "" {
		configFile, _ = homedir.Expand("~/.lacework.toml")
	}
	f, err := os.Open(configFile)
	if err == nil {
		defer f.Close()
		dec := toml.NewDecoder(f)
		err = dec.Decode(&laceworkProfiles)
		if err != nil {
			log.Warnf("Could not decode lacework profiles from {primary:%s} - {warning:%s}", configFile, err)
		} else {
			log.Debugf("Loaded lacework config file {primary:%s}", configFile)
		}
		for n, p := range laceworkProfiles {
			p.Name = n
		}
	} else {
		laceworkProfiles = nil
	}
}

func getLaceworkProfile(name string) *LaceworkProfile {
	if laceworkProfiles == nil {
		LoadLaceworkProfiles("")
	}
	return laceworkProfiles[name]
}

func GetDefaultLaceworkProfile() *LaceworkProfile {
	if laceworkProfiles == nil {
		LoadLaceworkProfiles("")
	}
	if len(laceworkProfiles) == 1 {
		for _, p := range laceworkProfiles {
			return p
		}
	}
	return laceworkProfiles["default"]
}

func GetLaceworkProfiles() (result []*LaceworkProfile) {
	if laceworkProfiles == nil {
		LoadLaceworkProfiles("")
	}
	for _, p := range laceworkProfiles {
		result = append(result, p)
	}
	return
}
