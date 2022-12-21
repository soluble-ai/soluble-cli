package configure

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T, assertions func(r *http.Request)) {
	t.Helper()
	httpmock.Activate()
	t.Cleanup(httpmock.Deactivate)
	httpmock.ZeroCallCounters()
	httpmock.RegisterResponder("GET", "https://api.soluble.cloud/api/v1/users/profile",
		func(r *http.Request) (*http.Response, error) {
			response, err := httpmock.NewJsonResponse(200, httpmock.File("testdata/profile.json"))
			assertions(r)
			return response, err
		})
}

func TestAuthWithEnv(t *testing.T) {
	assert := assert.New(t)
	t.Setenv("LW_COMPONENT_NAME", "iac")
	t.Setenv("LW_ACCOUNT", "test")
	t.Setenv("LW_IAC_ORGANIZATION", "123456789012")
	t.Setenv("LW_API_TOKEN", "12345")
	t.Setenv("SOLUBLE_API_SERVER", "https://api.soluble.cloud")
	setupTest(t, func(r *http.Request) {
		assert.Equal("test.lacework.net", r.Header.Get("X-LW-Domain"))
		assert.Equal("Token 12345", r.Header.Get("X-LW-Authorization"))
		assert.Equal("123456789012", r.Header.Get("X-SOLUBLE-ORG-ID"))
	})
	// In this case all information comes from the environment and there's
	// no need to run configure
	opts := &options.ClientOpts{}
	api, err := opts.GetAPIClient()
	if !assert.NoError(err) {
		return
	}
	httpmock.ActivateNonDefault(api.GetClient().GetClient())
	if !assert.NoError(err) {
		return
	}
	assert.Equal("123456789012", api.Organization)
	assert.Equal("12345", api.LaceworkAPIToken)
	n, err := api.Get("/api/v1/users/profile")
	if assert.NoError(err) {
		assert.Equal(1, httpmock.GetTotalCallCount())
		assert.Equal("123456789012", n.Path("data").Path("currentOrgId").AsText())
	}
}

func TestConfigure(t *testing.T) {
	assert := assert.New(t)
	t.Setenv("SOLUBLE_API_SERVER", "")
	setupTest(t, func(r *http.Request) {
		assert.Equal("test.lacework.net", r.Header.Get("X-LW-Domain"))
		assert.Equal("Token 12345", r.Header.Get("X-LW-Authorization"))
		assert.Equal("123456789012", r.Header.Get("X-SOLUBLE-ORG-ID"))
	})
	httpmock.RegisterResponder("POST", "https://test.lacework.net/api/v2/access/tokens",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"token":     "12345",
			"expiresAt": time.Now().Add(5 * time.Minute),
		}))
	configDir := t.TempDir()
	t.Setenv("SOLUBLE_CONFIG_DIR", configDir)
	config.Reset()
	config.Load()
	t.Cleanup(config.Reset)
	config.LoadLaceworkProfiles("testdata/lacework.toml")
	t.Cleanup(func() { config.LoadLaceworkProfiles("") })
	assert.NotNil(config.GetDefaultLaceworkProfile())
	configCommand := &configureCommand{}
	configCommand.APIConfig.Organization = "123456789012"
	configCommand.clientHook = func(c *api.Client) {
		httpmock.ActivateNonDefault(c.GetClient().GetClient())
	}
	n, err := configCommand.Run()
	assert.NoError(err)
	assert.NotNil(n)
	assert.Equal(2, httpmock.GetTotalCallCount())
	dat, err := os.ReadFile(filepath.Join(configDir, "iac-config.json"))
	assert.NoError(err)
	n, _ = jnode.FromJSON(dat)
	assert.Equal(1, n.Path("Profiles").Size())
	c := n.Path("Profiles").Path("default")
	assert.Equal("123456789012", c.Path("Organization").AsText())
	assert.Equal("test", c.Path("LaceworkProfileName").AsText())
}
