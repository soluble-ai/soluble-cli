package download

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
)

func isLatestTag(tag string) bool {
	return tag == "" || tag == "latest"
}

func getGithubReleaseAsset(owner, repo, tag string) (*github.RepositoryRelease, *github.ReleaseAsset, error) {
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
	for _, asset := range assets {
		name := asset.GetName()
		if isThisRuntimeRelease(name) {
			return release, asset, nil
		}
	}
	return nil, nil, fmt.Errorf("cannot find release for %s/%s", owner, repo)
}

var archNames = map[string][]string{
	"amd64": {"x86_64"},
	"386":   {"x86", "i386"},
}

func isThisRuntimeRelease(r string) bool {
	r = strings.ToLower(r)
	if !strings.Contains(r, "_"+runtime.GOOS+"_") {
		return false
	}
	if strings.Contains(r, "_"+runtime.GOARCH) {
		return true
	}
	names := archNames[runtime.GOARCH]
	for _, name := range names {
		if strings.Contains(r, "_"+name) {
			return true
		}
	}
	return false
}
