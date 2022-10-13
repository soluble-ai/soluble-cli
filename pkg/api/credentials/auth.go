package credentials

import (
	"fmt"
	"os"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/config"
)

func GetDomain(account string) string {
	if !strings.HasSuffix(account, ".lacework.net") {
		return fmt.Sprintf("%s.lacework.net", account)
	}
	return account
}

func ConfigureLaceworkAuth(cfg *api.Config) error {
	laceworkProfile := config.Config.Lacework
	if cfg.Domain == "" {
		account := os.Getenv("LW_ACCOUNT")
		if account == "" && laceworkProfile != nil {
			account = laceworkProfile.Account
		}
		if account != "" {
			cfg.Domain = GetDomain(account)
		}
	}
	if cfg.Domain == "" {
		// If we don't have the domain then we won't use lacework
		// authentication
		return nil
	}
	if cfg.LaceworkAPIToken == "" {
		cfg.LaceworkAPIToken = os.Getenv("LW_API_TOKEN")
	}
	if cfg.LaceworkAPIToken == "" {
		creds := Load()
		pc := creds.Find(config.Config.ProfileName)
		if pc.IsNearExpiration() {
			if laceworkProfile != nil {
				if err := pc.RefreshToken(GetDomain(laceworkProfile.Account), laceworkProfile.APIKey, laceworkProfile.APISecret); err != nil {
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
