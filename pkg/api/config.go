package api

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type Config struct {
	Organization               string
	OrganizationFromConfigFile bool
	LegacyAPIToken             string
	LaceworkAccount            string
	LaceworkAPIToken           string
	LaceworkAPIKey             string
	LaceworkAPISecret          string
	APIServer                  string
	APIPrefix                  string
	Debug                      bool
	TLSNoVerify                bool
	Timeout                    time.Duration
	RetryCount                 int
	RetryWaitSeconds           float64
	Headers                    []string
}

func (c *Config) SetValues() *Config {
	if config.IsRunningAsComponent() {
		// We unconditionally use the values from the lacework CLI
		c.LaceworkAPIToken = os.Getenv("LW_API_TOKEN")
		c.LaceworkAccount = os.Getenv("LW_ACCOUNT")
		c.LaceworkAPIKey = os.Getenv("LW_API_KEY")
		c.LaceworkAPISecret = os.Getenv("LW_API_SECRET")
		log.Debugf("Running as component and using LW_ACCOUNT={primary:%s} from environment",
			c.LaceworkAccount)
	} else if lwp := config.Get().GetLaceworkProfile(); lwp != nil {
		c.LaceworkAPIKey = lwp.APIKey
		c.LaceworkAPISecret = lwp.APISecret
		log.Debugf("Using lacework api key and secret from linked lacework profile {primary:%s}",
			lwp.Name)
		if c.LaceworkAccount == "" {
			c.LaceworkAccount = lwp.Account
			log.Debugf("Using account {primary:%s} linked to lacework profile {primary:%s}",
				c.LaceworkAccount, lwp.Name)
		}
	} else if c.LegacyAPIToken == "" {
		c.LegacyAPIToken = os.Getenv("SOLUBLE_API_TOKEN")
		if c.LegacyAPIToken == "" {
			c.LegacyAPIToken = config.Get().APIToken
		}
	}
	if c.APIServer == "" {
		c.APIServer = os.Getenv("LW_IAC_API_URL")
		if c.APIServer == "" {
			c.APIServer = os.Getenv("SOLUBLE_API_SERVER")
			if c.APIServer == "" {
				c.APIServer = config.Get().APIServer
			}
		}
	}
	if c.Organization == "" {
		c.Organization = os.Getenv("LW_IAC_ORGANIZATION")
		if c.Organization == "" {
			c.Organization = config.Get().Organization
			if c.Organization != "" {
				c.OrganizationFromConfigFile = true
			}
		}
	}
	return c
}

func (c *Config) ensureLaceworkAPIToken() error {
	if c.LaceworkAccount != "" && c.LaceworkAPIToken == "" {
		creds := loadCredentials()
		pc := creds.Find(config.Get().ProfileName)
		if pc.IsNearExpiration() {
			if c.LaceworkAPIKey != "" && c.LaceworkAPISecret != "" {
				if err := pc.RefreshToken(c.GetDomain(), c.LaceworkAPIKey, c.LaceworkAPISecret); err != nil {
					return err
				}
				if err := creds.Save(); err != nil {
					return err
				}
			}
		}
		c.LaceworkAPIToken = pc.Token
	}
	return nil
}

func (c *Config) GetAppURL() string {
	const httpAPI = "https://api."
	apiServer := c.APIServer
	if strings.HasPrefix(apiServer, httpAPI) {
		return "https://app." + apiServer[len(httpAPI):]
	}
	return "https://app.soluble.cloud"
}

func (c *Config) GetDomain() string {
	if !strings.HasSuffix(c.LaceworkAccount, ".lacework.net") {
		return fmt.Sprintf("%s.lacework.net", c.LaceworkAccount)
	}
	return c.LaceworkAccount
}

// Verify that the configuration chosen is usuable
// In particular, verify that LW_ACCOUNT is the same, because otherwise Organization
// is likely to be wrong.
func (c *Config) Validate() error {
	// Try and detect a couple of situations where the IAC component
	// is not configured properly and tell the user what to do
	if !config.IsRunningAsComponent() {
		if c.LegacyAPIToken != "" && c.LaceworkAccount == "" {
			log.Infof("Using legacy soluble authentication")
			return nil
		}
	}
	configuredAccount := config.Get().ConfiguredAccount
	laceworkProfileName := config.Get().LaceworkProfileName
	profileName := config.Get().ProfileName
	if config.IsRunningAsComponent() && config.Get().APIToken != "" {
		log.Errorf("The IAC profile {info:%s} can only be used when directly invoking soluble and not when running with the Lacework CLI",
			profileName)
		return fmt.Errorf("cannot use 'lacework iac ...' with this profile")
	}
	if c.LegacyAPIToken == "" && c.LaceworkAccount == "" {
		if configuredAccount != "" {
			log.Errorf("The IAC profile {info:%s} cannot be used when invoking the IAC component directly.",
				profileName)
			log.Infof("Run the command using the lacework CLI instead with {primary:lacework iac ...}")
			return fmt.Errorf("must use 'lacework iac ...'")
		}
		// We have no configuration at all.
		if config.IsRunningAsComponent() {
			log.Infof("The IAC profile {info:%s} must be configured by running {primary:%s configure}",
				profileName, config.CommandInvocation())
			return fmt.Errorf("configuration required")
		} else {
			log.Infof("The IAC profile {info:%s} is not authenticated.  Run {primary:%s login} to setup.",
				profileName, config.CommandInvocation())
			return fmt.Errorf("login required")
		}
	}
	if c.LaceworkAccount != "" {
		// We've got a lacework account, so we either need to be linked to a
		// lacework profile (and that profile must exist), or the account
		// must match the account we've been configured with.
		if config.Get().LaceworkProfileName != "" {
			if config.IsRunningAsComponent() {
				// Setting LaceworkProfileName is only to be used when invoked directly
				log.Errorf("The IAC profile {info:%s} must only be used for direct invocation", profileName)
				return fmt.Errorf("this IAC profile cannot be used with the lacework CLI")
			}
			lwp := config.Get().GetLaceworkProfile()
			if lwp == nil {
				log.Errorf("The IAC profile {info:%s} is configured to use the lacework profile {primary:%s}",
					profileName, laceworkProfileName)
				log.Errorf("but that profile does not exist.")
				return fmt.Errorf("lacework profile %s does not exist", laceworkProfileName)
			}
			if c.LaceworkAccount != "" && c.LaceworkAccount != lwp.Account {
				log.Errorf("The IAC profile {info:%s} is configured to use account {info:%s} but is running with account {info:%s}",
					profileName, lwp.Account, c.LaceworkAccount)
				log.Infof("Run {primary:%s configure reconfigure} to change configuration, or",
					config.CommandInvocation())
				log.Infof("use {primary:%s configure switch-profile <new-name>} to create a new profile for this account",
					config.CommandInvocation())
				return fmt.Errorf("configuration is invalid")
			}
			log.Debugf("IAC profile {info:%s} is linked to lacework profile {info:%s}",
				profileName, laceworkProfileName)
			configuredAccount = c.LaceworkAccount
		}
		if c.OrganizationFromConfigFile {
			// If the Organization comes from a config file, then the config
			// file must match the account
			if laceworkProfileName == "" && c.LaceworkAccount != configuredAccount {
				log.Errorf("The IAC profile {info:%s} is configured to use account {info:%s} but is running with {info:%s}",
					profileName, configuredAccount, c.LaceworkAccount)
				log.Errorf("Run {primary:%s configure reconfigure} to change configuration, or use the {primary:--iac-profile <profile>} to use a different profile",
					config.CommandInvocation(), config.CommandInvocation())
				return fmt.Errorf("configuration is invalid")
			}
		}
		log.Debugf("Using lacework authentication for account {info:%s}", c.LaceworkAccount)
		return nil
	}
	return nil
}
