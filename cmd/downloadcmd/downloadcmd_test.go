package downloadcmd

import "testing"

func TestGithubRepo(t *testing.T) {
	owner, repo := githubRepo("github.com/soluble-ai/soluble-cli")
	if owner != "soluble-ai" || repo != "soluble-cli" {
		t.Error(owner, repo)
	}
}
