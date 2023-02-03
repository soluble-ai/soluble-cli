package opal

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/opal"

	"github.com/soluble-ai/soluble-cli/pkg/policy/manager"
	"github.com/stretchr/testify/assert"
)

func TestPolicies(t *testing.T) {
	assert := assert.New(t)
	m := InitM(t)
	err := m.DetectPolicy("testdata/passing/policies")
	assert.NoError(err)
	assert.Equal(1, len(m.Policies))
	tm, err := m.TestPolicies()
	assert.NoError(err)
	assert.Equal(0, tm.Failed)
	// Ensure we get all results
	assert.Equal(3, tm.Passed)
}

func TestPoliciesFail(t *testing.T) {
	assert := assert.New(t)
	m := InitM(t)
	err := m.DetectPolicy("testdata/failing/policies")
	assert.NoError(err)
	assert.Equal(1, len(m.Policies))
	tm, err := m.TestPolicies()
	assert.Error(err)
	assert.Equal(1, tm.Failed)
	assert.Equal(1, tm.Passed)
}

func TestGetCustomPoliciesDir204(t *testing.T) {
	assert := assert.New(t)
	apiConfig := &api.Config{
		APIServer:        "https://api.test",
		APIPrefix:        "/api/v1",
		Organization:     "1234",
		LegacyAPIToken:   "token",
		LaceworkAPIToken: "token",
	}

	o := &tools.AssessmentOpts{
		ToolOpts: tools.ToolOpts{
			RunOpts: tools.RunOpts{
				PrintClientOpts: options.PrintClientOpts{
					ClientOpts: options.ClientOpts{
						APIConfig: *apiConfig,
					},
				},
				ToolPath: getToolPath(t),
			},
			Tool: &opal.Tool{},
		},
	}

	client, _ := o.RunOpts.ClientOpts.GetAPIClient()
	httpmock.ActivateNonDefault(client.GetClient().GetClient())
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(http.MethodGet,
		"https://api.test/api/v1/org/1234/policies/opal/policies.zip",
		httpmock.NewBytesResponder(http.StatusNoContent, []byte{}))

	customPoliciesDir, err := o.GetCustomPoliciesDir("opal")

	assert.Equal(1, httpmock.GetTotalCallCount())
	assert.NoError(err)
	assert.Equal(customPoliciesDir, "")
}

func InitM(t *testing.T) *manager.M {
	toolPath := getToolPath(t)
	return &manager.M{RunOpts: tools.RunOpts{ToolPath: toolPath}}
}

func getToolPath(t *testing.T) string {
	toolPath, ok := os.LookupEnv("TEST_OPAL_TOOL_PATH")
	if ok {
		// TEST_OPAL_TOOL_PATH should be set to the binary location under the opal repo to run the tests with a local opal binary
		log.Infof("TEST_OPAL_TOOL_PATH=%s", toolPath)
		assert.True(t, strings.HasSuffix(toolPath, "bin/opal"), "env var %s=%s is not a valid opal executable path", "TEST_OPAL_TOOL_PATH", toolPath)
	}
	return toolPath
}
