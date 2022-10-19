package credentials

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml/v2"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

// Credentials are saved in ~/.config/cli-credentials in toml format
// very much like the AWS cli does.  The file is updated atomically, but
// with no provision for coordinating parallel invocations of the CLI.

type ProfileCredentials struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type Credentials map[string]*ProfileCredentials

var (
	credentials Credentials
)

func Load() Credentials {
	if credentials != nil {
		return credentials
	}
	path, err := getCredentialsPath()
	if err != nil {
		log.Warnf("Could not find credentials - {warning:%s}", err)
	} else {
		dat, err := os.ReadFile(path)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Warnf("Could not read credentials - {warning:%s}", err)
			}
		} else {
			if err := toml.Unmarshal(dat, &credentials); err != nil {
				log.Warnf("Could not decode credentials - {warning:%s}", err)
			}
		}
	}
	if credentials == nil {
		credentials = Credentials{}
	}
	return credentials
}

func getCredentialsPath() (string, error) {
	credentialsDir := os.Getenv("SOLUBLE_CONFIG_DIR")
	if credentialsDir != "" {
		return filepath.Join(credentialsDir, "cli-credentials"), nil
	}
	return homedir.Expand("~/.config/lacework/cli-credentials")
}

func (c Credentials) Find(profileName string) *ProfileCredentials {
	pc := c[profileName]
	if pc == nil {
		pc = &ProfileCredentials{}
		c[profileName] = pc
	}
	return pc
}

func (c Credentials) Save() error {
	if len(credentials) == 0 {
		return nil
	}
	path, err := getCredentialsPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, "cli-credentials*")
	if err != nil {
		return err
	}
	enc := toml.NewEncoder(temp)
	if err = enc.Encode(credentials); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	log.Debugf("Saving credentials to {primary:%s}", path)
	return os.Rename(temp.Name(), path)
}

func (p *ProfileCredentials) IsNearExpiration() bool {
	return p.Token == "" || p.ExpiresAt.Before(time.Now().Add(-30*time.Second))
}

func (p *ProfileCredentials) RefreshToken(domain string, keyID string, secretKey string) error {
	url := fmt.Sprintf("https://%s/api/v2/access/tokens", domain)
	log.Debugf("Refreshing auth token from {info:%s}", url)
	body := jsonBody(map[string]interface{}{
		"keyId":      keyID,
		"expiryTime": 3600,
	})
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", api.UserAgent)
	req.Header.Set("X-LW-UAKS", secretKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := api.RClient.GetClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("could not get access token from %s - %s", url, resp.Status)
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(respBytes, p); err != nil {
		return fmt.Errorf("could not decode access token from %s", url)
	}
	return nil
}

func jsonBody(m map[string]interface{}) *bytes.Buffer {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	_ = enc.Encode(m)
	return buf
}
