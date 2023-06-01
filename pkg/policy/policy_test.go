package policy

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockPolicyType struct{}

func (mockPolicyType) GetName() string {
	return "mockPolicyType"
}
func (mockPolicyType) GetCode() string {
	return "mock"
}

func (t mockPolicyType) PreparePolicies(policies []*Policy, dest string) error {
	return nil
}

func TestResolvePolicyIDOkay(t *testing.T) {
	assertPolicyIDOkay(t, newTestStore(), "a_test_policy", "c-mock-a-test-policy")
}

func TestResolvePolicyIDExists(t *testing.T) {
	store := newTestStore()
	path := "a_test_policy"
	assertPolicyIDOkay(t, store, path, "c-mock-a-test-policy")
	policyID, err := store.resolvePolicyID(mockPolicyType{}, path)
	assert.Equal(t, "", policyID)
	assert.Error(t, err)
	assert.Equal(t, errors.New("a policy with id: c-mock-a-test-policy already exists"), err)
}

func TestLoadSinglePolicyCustom(t *testing.T) {
	store := NewDownloadStore("testdata")
	policy, err := store.LoadSinglePolicy(GetPolicyType("opal"), "testdata/downloadedpolicies/opal/downloaded_custom_policy")
	assert.NoError(t, err)
	assert.Equal(t, policy.ID, "c-opl-downloaded-custom-policy")
}

func TestLoadSinglePolicyLacework(t *testing.T) {
	store := NewDownloadStore("testdata")
	policy, err := store.LoadSinglePolicy(GetPolicyType("opal"), "testdata/downloadedpolicies/opal/downloaded_iac_aws_iam_1")
	assert.NoError(t, err)
	assert.Equal(t, policy.ID, "lacework-downloaded-iac-aws-iam-1")
}

func assertPolicyIDOkay(t *testing.T, store *Store, path string, expected string) {
	policyID, err := store.resolvePolicyID(mockPolicyType{}, path)
	assert.Equal(t, expected, policyID)
	assert.Nil(t, err)
}

func newTestStore() *Store {
	return &Store{
		PolicyIds: make(map[string]string),
	}
}
