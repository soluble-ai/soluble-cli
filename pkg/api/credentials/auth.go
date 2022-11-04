package credentials

import (
	"fmt"
	"os"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

func GetDomain(account string) string {
	if !strings.HasSuffix(account, ".lacework.net") {
		return fmt.Sprintf("%s.lacework.net", account)
	}
	return account
}

func ConfigureLaceworkAuth(cfg *api.Config) error {
	if cfg.Domain == "" {
		account := config.Config.GetLaceworkAccount()
		if account != "" {
			cfg.Domain = GetDomain(account)
		}
	}
	if cfg.Domain == "" {
		// If we don't have the domain then we won't use lacework
		// authentication
		return nil
	}
	log.Infof("Using lacework authentication for {info:%s} and IAC org {info:%s}", cfg.Domain,
		cfg.Organization)
	if cfg.LaceworkAPIToken == "" {
		cfg.LaceworkAPIToken = os.Getenv("LW_API_TOKEN")
	}
	if cfg.LaceworkAPIToken == "" {
		creds := Load()
		pc := creds.Find(config.Config.ProfileName)
		if pc.IsNearExpiration() {
			apiKey := config.Config.GetLaceworkAPIKey()
			apiSecret := config.Config.GetLaceworkAPISecret()
			if apiKey != "" && apiSecret != "" {
				if err := pc.RefreshToken(cfg.Domain, apiKey, apiSecret); err != nil {
					return err
				}
				if err := creds.Save(); err != nil {
					return err
				}
			}
		}
		cfg.LaceworkAPIToken = pc.Token
	}
	return nil
}
