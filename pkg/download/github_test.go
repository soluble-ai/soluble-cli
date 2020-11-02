package download

import "testing"

func TestParseGithubRepo(t *testing.T) {
	owner, repo := parseGithubRepo("github.com/soluble-ai/soluble-cli")
	if owner != "soluble-ai" || repo != "soluble-cli" {
		t.Error(owner, repo)
	}
}
