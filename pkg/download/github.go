package download

import (
	"context"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
)

func isLatestTag(tag string) bool {
	return tag == "" || tag == "latest"
}

func parseGithubRepo(url string) (string, string) {
	const githubCom = "github.com"
	if strings.HasPrefix(url, githubCom) {
		parts := strings.Split(url[len(githubCom)+1:], "/")
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	return "", ""
}

func getGithubReleaseAsset(owner, repo, tag string, releaseMatcher GithubReleaseMatcher) (*github.RepositoryRelease, *github.ReleaseAsset, error) {
	client := github.NewClient(nil)
	var release *github.RepositoryRelease
	var err error
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	if isLatestTag(tag) {
		release, _, err = client.Repositories.GetLatestRelease(ctx, owner, repo)
	} else {
		release, _, err = client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	}
	if err != nil {
		return nil, nil, err
	}
	assets, _, err := client.Repositories.ListReleaseAssets(ctx, owner, repo, release.GetID(), nil)
	if err != nil {
		return nil, nil, err
	}
	asset, err := chooseReleaseAsset(assets, releaseMatcher)
	if err != nil {
		return nil, nil, err
	}
	return release, asset, nil
}
