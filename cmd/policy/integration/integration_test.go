//go:build integration

package integration

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
)

func TestPolicyVet(t *testing.T) {
	vet := test.NewCommand(t, "early-access", "policy", "vet", "-d", "../../../pkg/policy/testdata")
	vet.Must(vet.Run())
}
